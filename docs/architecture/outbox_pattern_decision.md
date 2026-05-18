# Outbox Pattern 採用理由

> 「仮登録は成功したのにメールが届かない」を防ぐ設計。仮登録レコードとメール送信ジョブを同一トランザクションで作成し、SMTP 送信は worker が非同期に行う。SMTP 直送では 2 種類の不整合が発生し、どちらも自動回復できない。諦めたこと: メールの瞬時配信（送信は API レスポンス後 1〜2 秒の遅延）。

## Problem

会員登録フローでは「仮登録レコードのDB書き込み」と「確認メールの送信」を確実に対で実行する必要がある。
この2操作をどう結合するかで、以下の事故が発生する。

### 事故パターン A — SMTP直送（トランザクション外）

```
BEGIN
  INSERT user_registration_requests
COMMIT
          ↓
    SMTP送信 ← ここで失敗（SMTPサーバ停止・ネットワーク断）
```

- DBには仮登録レコードが存在する
- ユーザにメールは届かない
- どのレコードが未送信かを後から特定する手段がない
- workerも存在しないため、手動で再送するしかない

### 事故パターン B — SMTP直送（トランザクション内）

```
BEGIN
  INSERT user_registration_requests
  SMTP送信 ← 成功。メールは送られた
ROLLBACK  ← その後の別の処理が失敗
```

- メールはすでに届いている
- DBにはレコードが残らない
- ユーザはメール内のURLをクリックするが、対応するtoken_hashがDBに存在せず本登録できない
- SMTPはROLLBACKの概念を持たないため、送信済みメールを取り消せない

### 事故パターン C — APIがSMTPレイテンシを直接受ける

- SMTP送信には通常100〜2000ms、障害時はタイムアウトまで数十秒かかる
- SMTP直送にするとAPIのレスポンスタイムがSMTPサーバの状態に依存する
- SMTPサーバが重い・停止中のとき、APIリクエストが全部タイムアウトする
- 仮登録エンドポイントが「使えない状態」になる

## Decision

`mail_outboxes` テーブルを介したOutbox Patternを採用する。

```
BEGIN
  INSERT INTO user_registration_requests  ← 仮登録
  INSERT INTO mail_outboxes (            ← 送信ジョブ
    status='pending',
    payload='{"email":..., "url":..., "lang":...}',
    next_attempt_at=NOW()
  )
COMMIT
       ↓（トランザクション成功後）
worker が1秒間隔でpolling
  → FetchPending（status='pending' AND next_attempt_at <= NOW()）
  → status を 'processing' に更新（FOR UPDATE SKIP LOCKED）
  → SMTP送信
  → MarkSent（status='sent', payload='{}'）
```

DBへの2つのINSERTを同一トランザクションに含めることで、原子性を確保する。
「仮登録レコードが存在する = 送信ジョブも存在する」が保証される。
APIはSMTP送信を待たずに201を返す。

### なぜpollingか

- Redisのpub/subやKafkaなどのメッセージキューを使う場合、DBへの書き込みとキューへのenqueueが別操作になる。トランザクションでまとめられないため、DB成功・キュー失敗の場合にメールが送られないまま失われる
- DBだけで完結するpollingはインフラの追加が不要で、障害点が少ない
- 確認メールは「瞬時」である必要がない。1〜2秒の遅延はユーザ体験上問題にならない
- `FOR UPDATE SKIP LOCKED` により複数workerが同一レコードを取得しない

### なぜpayload clearが必要か

`mail_outboxes.payload` には本登録URL（平文トークンを含む）が格納される。

```json
{
  "email": "user@example.com",
  "url": "https://example.com/registration/verify?token=dGhpcyBpcyBhIHRlc3QgdG9rZW4",
  "lang": "ja"
}
```

送信後もpayloadを保持し続けると、DBダンプ・バックアップ漏洩時に「期限内の本登録URL」が一括流出するリスクがある。
`token_hash` は SHA-256 で保存しているのに対し、payloadの平文URLからは生トークンが即座に取得できる。

MarkSent時に `payload = '{}'` で上書きすることで、平文URLのDB残留期間を「SMTP送信完了まで」に限定する。

## Alternatives Considered

| 案 | 却下理由 |
|---|---|
| SMTP直送（トランザクション外） | 送信失敗時に未送信レコードの特定・再送手段がない |
| SMTP直送（トランザクション内） | SMTPはROLLBACK不可。送信済みメールとDBが不整合になる |
| Redis pub/sub | DBとキューへの書き込みが2操作になり原子性が失われる。インフラ追加が必要 |
| SQS / Kafka | 同上。現時点のスケールでは過剰 |
| payload不要論（workerがDBからurl再構築） | workerが`token_hash`から平文トークンを逆算できないため不可。URL構築には生トークンが必要 |

## Consequences

- **at-least-once配信になる**: MarkSentがDB書き込み失敗で成功しない場合、同一メールが再送される可能性がある。詳細は `retry_strategy_decision.md` と `at_least_once_and_idempotency.md` を参照
- **pollingの遅延**: メール送信はAPIレスポンス後1〜2秒後になる。「送信ボタンを押した瞬間に届く」ではない
- **payload clearのタイミング依存**: workerがSMTP送信成功後・MarkSent呼び出し前にcrashすると、次のstuck recovery（デフォルト15分後）まで平文URLがDBに残る。この15分間はDBダンプ漏洩時のリスクウィンドウになる
- **workerが停止していても仮登録は成功する**: APIはworkerの生死を感知しない。workerが長時間停止すると、ユーザはメールが届かないと感じるが、再起動後に自動送信される

## Why This Matters Later

オークションのホールド通知・落札通知・支払い完了通知など、今後増えるすべての非同期通知に同じ基盤を流用する。
Outbox Patternのインフラを会員登録フェーズで固めることで、通知系機能の実装コストが下がる。

また、「支払い確認メール」や「入札更新通知」では重複送信の許容度が低い可能性がある。
その場合は idempotency key + mail_outboxes への UNIQUE 制約の追加で対応するが、基盤はそのまま使える。
