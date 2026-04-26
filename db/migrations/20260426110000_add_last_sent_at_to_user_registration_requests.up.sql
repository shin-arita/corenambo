ALTER TABLE user_registration_requests
    ADD COLUMN last_sent_at TIMESTAMPTZ;

COMMENT ON COLUMN user_registration_requests.last_sent_at IS '最終メール送信日時';

CREATE INDEX idx_user_registration_requests_last_sent_at
    ON user_registration_requests(last_sent_at);
