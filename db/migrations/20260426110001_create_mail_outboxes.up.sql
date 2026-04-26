CREATE TABLE IF NOT EXISTS mail_outboxes (
        id uuid PRIMARY KEY,
        mail_type varchar(100) NOT NULL,
        to_email varchar(255) NOT NULL,
        payload jsonb NOT NULL,
        status varchar(20) NOT NULL DEFAULT 'pending',
        retry_count integer NOT NULL DEFAULT 0,
        next_attempt_at timestamptz NOT NULL,
        sent_at timestamptz,
        last_error text,
        created_at timestamptz NOT NULL,
        updated_at timestamptz NOT NULL,
        CONSTRAINT mail_outboxes_status_check CHECK (status IN ('pending', 'processing', 'sent', 'failed')),
        CONSTRAINT mail_outboxes_retry_count_check CHECK (retry_count >= 0)
);

CREATE INDEX IF NOT EXISTS idx_mail_outboxes_status_next_attempt_at
        ON mail_outboxes (status, next_attempt_at);

CREATE INDEX IF NOT EXISTS idx_mail_outboxes_created_at
        ON mail_outboxes (created_at);

COMMENT ON TABLE mail_outboxes IS 'メール送信アウトボックス（非同期送信キュー）';

COMMENT ON COLUMN mail_outboxes.id IS '主キー（UUIDv7）';
COMMENT ON COLUMN mail_outboxes.mail_type IS 'メール種別（例：user_registration）';
COMMENT ON COLUMN mail_outboxes.to_email IS '送信先メールアドレス';
COMMENT ON COLUMN mail_outboxes.payload IS 'メール本文生成用データ（JSON）';
COMMENT ON COLUMN mail_outboxes.status IS 'ステータス（pending / processing / sent / failed）';
COMMENT ON COLUMN mail_outboxes.retry_count IS 'リトライ回数';
COMMENT ON COLUMN mail_outboxes.next_attempt_at IS '次回送信予定時刻';
COMMENT ON COLUMN mail_outboxes.sent_at IS '送信成功時刻';
COMMENT ON COLUMN mail_outboxes.last_error IS '最終エラー内容';
COMMENT ON COLUMN mail_outboxes.created_at IS '作成日時';
COMMENT ON COLUMN mail_outboxes.updated_at IS '更新日時';
