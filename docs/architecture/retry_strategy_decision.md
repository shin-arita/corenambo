# リトライ戦略 — retryable / non-retryable / stuck

> worker の失敗を「SMTP 接続エラー」と「テンプレート壊れ」で同じ処理にすると、無限ループかメール永久不達のどちらかが起きる。失敗を 4 種類に分類し、それぞれ異なる遷移を取る。また、worker crash によって `status='processing'` のまま永続するレコード（stuck mail）を 15 分後に自動リカバリする設計を含む。諦めたこと: exponential backoff（5 分固定で十分と判断）。

## Problem

workerがSMTP送信を試みるとき、失敗の原因は大きく2種類ある。
この区別を誤ると、以下の事故が起きる。

### 事故パターン A — すべてリトライすると無限ループになる

テンプレートのparse失敗・payloadのURLが空など、コードやデータに起因する失敗は何度リトライしても成功しない。
上限なくリトライし続けると、workerのpollingサイクルを消費し続け、正常なレコードの処理が遅延する。

### 事故パターン B — すべてを即FailedにするとSMTP一時障害で登録が完結しない

SMTPサーバの一時停止・ネットワーク断など、数分後には回復する障害で即FailedにするとユーザにはSMTP回復後も確認メールが届かない。
仮登録は成功したのにメールが来ない状態が永続する。

### 事故パターン C — stuck mail（最も見えにくい）

```
worker: FetchPending → status を 'processing' に更新
worker: SMTP送信 → 成功
worker: MarkSent を呼び出す前に crash（OOM / SIGKILL / Dockerコンテナ再起動）
```

この状態でrecovery機構がないと、レコードは `status='processing'` のまま永続する。
`FetchPending` は `status='pending'` のレコードしか取得しないため、このレコードは二度とworkerに処理されない。
ユーザはメールが届いているのにそのことをシステムが認識しておらず、運用側からは「pending でも sent でもない謎のレコード」として残る。

## Decision

失敗を4種類に分類し、それぞれ異なる遷移を取る。

### 分類 1: retryable — SMTP接続エラー・タイムアウト

```
MarkRetry(
  status    = 'pending',
  retry_count = retry_count + 1,
  next_attempt_at = NOW() + 5分
)
```

- 一時的な障害として扱い、5分後に再試行する
- `next_attempt_at <= NOW()` の条件でFetchPendingが次のworkerサイクルで再取得する

### 分類 2: invalid payload — JSON parse 失敗

```
MarkFailed(
  status          = 'failed',
  retry_count     = 変更なし（最終状態なので増やさない）,
  next_attempt_at = NOW() + 5分
)
```

payloadのJSONパースに失敗した場合はコードまたはデータの問題であり、リトライしても解消しない。
即座にMarkFailedを呼ぶ。

`next_attempt_at` を5分後に設定するのは、`FetchPending` の取得条件（`next_attempt_at <= NOW()`）から外すためではなく、`status='failed'` の時点でFetchPendingには引っかからないため実運用上は意味を持たない。
調査・手動復旧・将来の再投入設計のために記録として残る。

### 分類 3: non-retryable — NonRetryableMailError

```
MarkFailed(
  status          = 'failed',
  retry_count     = 変更なし（最終状態なので増やさない）,
  next_attempt_at = NOW() + 24時間
)
```

non-retryable と判断する条件（`mail` パッケージが `NonRetryableMailError` 型で返す）:

| 条件 | 理由 |
|---|---|
| payload の `url` フィールドが空 | URLなしでメールを作れない。リトライしても同じ |
| メールテンプレートのparse失敗 | テンプレート自体が壊れている。コードを直すまで回復しない |
| メールテンプレートのexecute失敗 | 同上 |

workerは `errors.As` でこの型を検出し、MarkRetryの代わりにMarkFailedを呼ぶ。

### 分類 4: リトライ上限到達

```
if retry_count >= WorkerMaxRetryCount {
  MarkFailed(
    status          = 'failed',
    retry_count     = 変更なし,
    next_attempt_at = NOW() + 24時間
  )
}
```

retryable な失敗でも上限回数（`WorkerMaxRetryCount`）に達したら諦める。
SMTPサーバが長時間停止している場合、永遠にリトライし続けることを防ぐ。

### MarkFailed の共通挙動

`MarkFailed` は3分類すべてで `next_attempt_at` を更新する。`status='failed'` のレコードは `FetchPending`（`status='pending'` 条件）の取得対象にならないため、`next_attempt_at` の値は通常の送信フローでは使われない。
ただし調査・手動復旧・将来的な再投入設計（例: `status` を `pending` に戻すスクリプト）のときに「いつ頃から再試行してよいか」の目安として参照できる。

| 分類 | next_attempt_at |
|---|---|
| invalid payload (JSON parse失敗) | NOW() + 5分 |
| NonRetryableMailError | NOW() + 24時間 |
| リトライ上限到達 | NOW() + 24時間 |

### stuck recovery

processing状態で `updated_at < NOW() - 15分` のレコードを `status='pending'` に戻す。

```sql
UPDATE mail_outboxes
SET status = 'pending', next_attempt_at = NOW()
WHERE status = 'processing'
  AND updated_at < NOW() - INTERVAL '15 minutes'
```

- workerの起動時またはpollingループ内で定期実行する
- 15分はworkerの最大処理時間（正常なSMTP送信 + MarkSent）を大幅に超えた値
- これにより、worker crashによるstuck mailが最大15分後にpendingへ戻る

stuck recoveryによって復旧したレコードは再送される。
つまりユーザには「同じ確認メール」が届く可能性がある。これは意図的な許容（詳細は `at_least_once_and_idempotency.md` を参照）。

## Alternatives Considered

| 案 | 却下理由 |
|---|---|
| 失敗を区別せずすべてリトライ | non-retryable な失敗が無限ループになる |
| 失敗を区別せずすべて即Failed | SMTP一時障害で確認メールが永遠に届かない |
| exponential backoff | 5分固定で会員登録の確認メール用途には十分。指数バックオフの実装コストを正当化できない |
| stuck recovery不要 | worker crashが現実に起きるため不要にはできない。特にDockerコンテナ環境ではOOMによる再起動が起きる |
| stuck recoveryに専用cronを使う | pollingループと同一プロセスで実行することでインフラ追加を避ける |

## Consequences

- **重複送信のリスク**: stuck recoveryがレコードをpendingに戻すと、実際にはすでに送信済みのメールが再送される。ユーザは同じ確認メールを2通受け取る可能性がある（許容設計。詳細は `at_least_once_and_idempotency.md`）
- **SMTPサーバが長時間停止した場合のデータ蓄積**: pendingレコードがリトライ上限到達まで積み上がる。SMTP回復後に一斉送信が発生する可能性がある
- **stuck recovery までの最大15分**: worker crashから最大15分間、該当ユーザには「メールが来ない」状態が続く。これは意図的なトレードオフ（短すぎるタイムアウトは、正常な処理中のレコードを stuck 扱いするリスクがある）
- **`failed` レコードは自動復旧しない**: リトライ上限到達後は人手での確認・再キューイングが必要。運用監視の対象になる

## Why This Matters Later

支払い・入札・落札など、今後追加するすべてのworker処理に同じ失敗分類が必要になる。
会員登録の確認メールで失敗分類・stuck recovery・リトライ上限の設計を固めることで、後続機能で同じ議論をしなくて済む。

特に支払い確認メールは「届かない」ことの影響が大きいため、リトライとstuck recoveryが機能していることの信頼性が重要になる。
