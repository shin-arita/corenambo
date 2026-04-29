CREATE TABLE user_registration_requests (
                                            id UUID PRIMARY KEY,
                                            email VARCHAR(255) NOT NULL,
                                            token_hash VARCHAR(255) NOT NULL,
                                            expires_at TIMESTAMPTZ NOT NULL,
                                            verified_at TIMESTAMPTZ,
                                            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                            CONSTRAINT chk_user_registration_requests_email_not_blank
                                                CHECK (btrim(email) <> ''),

                                            CONSTRAINT chk_user_registration_requests_token_hash_not_blank
                                                CHECK (btrim(token_hash) <> ''),

                                            CONSTRAINT chk_user_registration_requests_expires_at_after_created_at
                                                CHECK (expires_at > created_at),

                                            CONSTRAINT chk_user_registration_requests_verified_at_after_created_at
                                                CHECK (verified_at IS NULL OR verified_at >= created_at)
);

COMMENT ON TABLE user_registration_requests IS 'ユーザ仮登録';

COMMENT ON COLUMN user_registration_requests.id IS 'ID';
COMMENT ON COLUMN user_registration_requests.email IS 'メールアドレス';
COMMENT ON COLUMN user_registration_requests.token_hash IS '認証トークンハッシュ';
COMMENT ON COLUMN user_registration_requests.expires_at IS '有効期限';
COMMENT ON COLUMN user_registration_requests.verified_at IS '認証日時';
COMMENT ON COLUMN user_registration_requests.created_at IS '作成日時';

CREATE UNIQUE INDEX uq_user_registration_requests_token_hash
    ON user_registration_requests(token_hash);

CREATE INDEX idx_user_registration_requests_email
    ON user_registration_requests(email);

CREATE INDEX idx_user_registration_requests_expires_at
    ON user_registration_requests(expires_at);

CREATE INDEX idx_user_registration_requests_verified_at
    ON user_registration_requests(verified_at);
