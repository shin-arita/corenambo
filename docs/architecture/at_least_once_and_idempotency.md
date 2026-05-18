# at-least-once delivery と冪等性方針

> SMTP プロトコルに「送信済みか確認する」機能はない。worker crash のタイミング次第で、送信済みメールが再送される可能性を排除できない。確認メールの重複は明示的に許容し、本登録（ユーザ作成）は 3 層防御で二重作成を防ぐ非対称な設計を取る。諦めたこと: 確認メールの exactly-once 配信。

## Problem

メール送信において「exactly-once（厳密に1回だけ送る）」を実現しようとすると、解決できない問題が残る。

### なぜexactly-onceは実現できないか

workerがSMTP送信に成功した後、MarkSentを呼ぶ前にcrashした場合を考える。

```
worker: SMTP送信 → 成功。メールはSMTPサーバに渡った
worker: MarkSent呼び出し → crash（DB接続断 / OOM / コンテナ再起動）
         ↓
レコードは status='processing' のまま
stuck recovery（15分後）でstatus='pending'に戻る
         ↓
workerが再度FetchPending → 同じレコードを取得
worker: SMTP送信 → 同じメールを再送
```

この問題を防ぐには「SMTP送信済みか否か」をSMTPサーバに問い合わせる必要があるが、SMTPプロトコルにはその仕組みがない。
送信済みかどうかを確認する手段がないため、workerは送信するしかない。

アプリケーション層でexactly-onceを実現するには分散2フェーズコミット（SMTP+DB間）が必要になるが、SMTPはその参加者になれない。

**何を諦めたか**: 確認メールのexactly-once配信。稀なworker crashシナリオで同じ確認メールが2通届く可能性を許容する。

## Decision

### メール配信: at-least-once を明示的に許容する

確認メールの重複送信を設計上の許容範囲とする。

**なぜ重複を許容できるか**:

- 確認メールは同一内容・同一URL・同一トークンを含む
- ユーザが先に届いたメールでURLをクリックして本登録を完了すると `verified_at` がセットされる
- その後に届いた2通目のメールのURLをクリックすると `USED_REGISTRATION_TOKEN` が返る（害なし）
- 「同じ確認メールが2通届いた」はユーザ体験上の微小なノイズであり、機能的な障害ではない
- メールの重複送信は仮登録フローの性質上、ユーザが能動的に再送操作をしても起きる（別ブラウザから再送など）

### ユーザ作成: トランザクション内で二重防御する

本登録（`POST /api/v1/user-registrations/verify`）への並行リクエストによる二重ユーザ作成は許容しない。

**防御 1: `FOR UPDATE` による排他ロック**

```sql
SELECT * FROM user_registration_requests
WHERE token_hash = $1
FOR UPDATE
```

同一トークンに対する並行リクエストはここで直列化される。先のリクエストがロックを取得し、後のリクエストはロック解放まで待機する。
先のリクエストが `verified_at` をセットしてCOMMITした後、後のリクエストがロックを取得すると、`verified_at IS NOT NULL` のチェックで `USED_REGISTRATION_TOKEN` を返す。

**防御 2: `user_emails` のUNIQUEインデックス**

```sql
CREATE UNIQUE INDEX user_emails_email_unique_idx ON user_emails (LOWER(email));
```

アプリ層の`FindByEmail`チェックと、DBのUNIQUE制約の両方でメール重複をブロックする。
アプリ層のチェックと実際のINSERTの間に別トランザクションが同じメールでINSERTしても、DB制約がブロックする。

**防御 3: トークン単一使用**

`verified_at IS NOT NULL` のレコードに対するverify操作は `USED_REGISTRATION_TOKEN` を返す。
1つのトークンで複数のユーザが作成されることはない。

### フロントエンド: URLからトークンを除去する

本登録画面表示後に `history.replaceState` でURLから `?token=...` を除去する。

```javascript
// tokenをURLに残すと起きる問題
// - ブラウザのバックボタンで同じURLに戻り、再度POSTが発生する
// - ユーザがURLをコピーしてシェアすると、受け取った相手が本登録を横取りできる
// - Refererヘッダでtokenが外部サービスに送信される
```

URLからトークンを除去することで、ブラウザ操作による意図しない再送を防ぐ。

## Alternatives Considered

| 案 | 却下理由 |
|---|---|
| exactly-once配信（SMTP deduplication） | SMTPプロトコルにdeduplication機能がない。実装不可能 |
| 送信前に「送信済みフラグ」をDBにセットする | フラグセット成功・SMTP送信失敗の場合、メールが届かないまま「送信済み」になる。検出困難 |
| Mailgun / SendGridなどのidempotency key | 外部メールサービス依存になる。Mailpitで開発できなくなる。現時点でのスコープ外 |
| verify API をGET（冪等リクエスト）に変更する | 本登録はDB書き込みを伴うためGETは不適切 |

## Consequences

- **確認メールが稀に2通届く**: worker crashが発生し、stuck recoveryが動いた場合。発生頻度はworkerの安定性依存
- **ユーザの混乱リスクは低い**: どちらのメールのURLでも本登録は完了できる。2通目は「既に登録済み」画面に遷移するだけ
- **本登録の重複はない**: FOR UPDATE + verified_atチェック + user_emails UNIQUE制約の3層防御により、同一トークンから複数ユーザが作成されることはない
- **モニタリングが必要**: `failed` 状態のmail_outboxesレコードが増えている場合、SMTPサーバに問題がある。アラートの対象にする必要がある

## Why This Matters Later

支払い・入札・落札など、今後追加する機能では「重複許容の可否」を機能ごとに明示的に決める必要がある。

| 機能 | 重複許容 | 理由 |
|---|---|---|
| 確認メール | 許容 | 同一URLが複数届いても害なし |
| 支払い完了通知メール | 要検討 | 重複は混乱を招く可能性あり |
| 課金処理 | 不可 | idempotency keyとDB UNIQUE制約が必須 |
| 入札確定 | 不可 | 同一入札の二重確定を防ぐためFOR UPDATEと状態遷移管理が必要 |

「at-least-onceで十分か・exactly-onceが必要か」を機能設計の起点に置くことで、実装漏れと過剰実装の両方を防ぐ。
