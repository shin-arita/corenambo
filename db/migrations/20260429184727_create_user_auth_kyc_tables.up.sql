-- ユーザ本体
CREATE TABLE users (
                       id uuid PRIMARY KEY,
                       display_name varchar(100) NOT NULL,
                       status varchar(20) NOT NULL DEFAULT 'active',
                       created_at timestamptz NOT NULL DEFAULT NOW(),
                       updated_at timestamptz NOT NULL DEFAULT NOW(),
                       CONSTRAINT users_status_check CHECK (status IN ('active', 'suspended', 'withdrawn'))
);

COMMENT ON TABLE users IS 'ユーザ本体を管理するテーブル';
COMMENT ON COLUMN users.id IS 'ユーザID。UUID v7をGo側で生成する';
COMMENT ON COLUMN users.display_name IS '画面表示用のユーザ名';
COMMENT ON COLUMN users.status IS 'ユーザ状態。active:有効、suspended:停止、withdrawn:退会';
COMMENT ON COLUMN users.created_at IS '作成日時';
COMMENT ON COLUMN users.updated_at IS '更新日時';

-- ユーザメールアドレス
CREATE TABLE user_emails (
                             id uuid PRIMARY KEY,
                             user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
                             email varchar(255) NOT NULL,
                             is_primary boolean NOT NULL DEFAULT false,
                             verified_at timestamptz,
                             created_at timestamptz NOT NULL DEFAULT NOW(),
                             updated_at timestamptz NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE user_emails IS 'ユーザのメールアドレスを管理するテーブル。ログインIDとして使用する';
COMMENT ON COLUMN user_emails.id IS 'ユーザメールID。UUID v7をGo側で生成する';
COMMENT ON COLUMN user_emails.user_id IS 'ユーザID';
COMMENT ON COLUMN user_emails.email IS 'メールアドレス。ログインIDとして使用する';
COMMENT ON COLUMN user_emails.is_primary IS '現在の主メールアドレスかどうか';
COMMENT ON COLUMN user_emails.verified_at IS 'メール認証完了日時。未認証の場合はNULL';
COMMENT ON COLUMN user_emails.created_at IS '作成日時';
COMMENT ON COLUMN user_emails.updated_at IS '更新日時';

CREATE UNIQUE INDEX user_emails_email_unique_idx
    ON user_emails (LOWER(email));

CREATE UNIQUE INDEX user_emails_primary_unique_idx
    ON user_emails (user_id)
    WHERE is_primary = true;

CREATE INDEX user_emails_user_id_idx
    ON user_emails (user_id);

-- ユーザパスワード
CREATE TABLE user_passwords (
                                user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE RESTRICT,
                                password_hash text NOT NULL,
                                password_updated_at timestamptz NOT NULL,
                                created_at timestamptz NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE user_passwords IS 'ユーザのパスワード情報を管理するテーブル';
COMMENT ON COLUMN user_passwords.user_id IS 'ユーザID';
COMMENT ON COLUMN user_passwords.password_hash IS 'パスワードハッシュ。平文パスワードは保存しない';
COMMENT ON COLUMN user_passwords.password_updated_at IS 'パスワード更新日時';
COMMENT ON COLUMN user_passwords.created_at IS '作成日時';

-- 本人確認
CREATE TABLE user_verifications (
                                    id uuid PRIMARY KEY,
                                    user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
                                    status varchar(20) NOT NULL,
                                    provider varchar(50) NOT NULL,
                                    provider_reference_id varchar(255),
                                    verification_method varchar(50) NOT NULL,
                                    document_type varchar(50) NOT NULL,
                                    submitted_at timestamptz,
                                    approved_at timestamptz,
                                    rejected_at timestamptz,
                                    rejection_reason text,
                                    created_at timestamptz NOT NULL DEFAULT NOW(),
                                    updated_at timestamptz NOT NULL DEFAULT NOW(),
                                    CONSTRAINT user_verifications_status_check CHECK (status IN ('pending', 'approved', 'rejected', 'expired', 'canceled')),
                                    CONSTRAINT user_verifications_document_type_check CHECK (document_type IN ('driver_license', 'my_number_card', 'passport', 'residence_card'))
);

COMMENT ON TABLE user_verifications IS 'ユーザの本人確認申請履歴を管理するテーブル';
COMMENT ON COLUMN user_verifications.id IS '本人確認ID。UUID v7をGo側で生成する';
COMMENT ON COLUMN user_verifications.user_id IS 'ユーザID';
COMMENT ON COLUMN user_verifications.status IS '本人確認状態。pending:申請中、approved:承認、rejected:却下、expired:期限切れ、canceled:取消';
COMMENT ON COLUMN user_verifications.provider IS '本人確認サービス提供元。自前保存ではなく外部eKYC利用を想定する';
COMMENT ON COLUMN user_verifications.provider_reference_id IS '外部eKYCサービス側の参照ID';
COMMENT ON COLUMN user_verifications.verification_method IS '本人確認方式';
COMMENT ON COLUMN user_verifications.document_type IS '本人確認書類種別';
COMMENT ON COLUMN user_verifications.submitted_at IS '申請日時';
COMMENT ON COLUMN user_verifications.approved_at IS '承認日時';
COMMENT ON COLUMN user_verifications.rejected_at IS '却下日時';
COMMENT ON COLUMN user_verifications.rejection_reason IS '却下理由';
COMMENT ON COLUMN user_verifications.created_at IS '作成日時';
COMMENT ON COLUMN user_verifications.updated_at IS '更新日時';

CREATE INDEX user_verifications_user_id_idx
    ON user_verifications (user_id);

CREATE INDEX user_verifications_user_status_created_idx
    ON user_verifications (user_id, status, created_at DESC);

CREATE UNIQUE INDEX user_verifications_provider_reference_unique_idx
    ON user_verifications (provider, provider_reference_id)
    WHERE provider_reference_id IS NOT NULL;

-- 本人確認イベント
CREATE TABLE user_verification_events (
                                          id uuid PRIMARY KEY,
                                          user_verification_id uuid NOT NULL REFERENCES user_verifications(id) ON DELETE RESTRICT,
                                          event_type varchar(100) NOT NULL,
                                          payload jsonb NOT NULL DEFAULT '{}'::jsonb,
                                          created_at timestamptz NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE user_verification_events IS '本人確認の状態変更イベントを管理するテーブル';
COMMENT ON COLUMN user_verification_events.id IS '本人確認イベントID。UUID v7をGo側で生成する';
COMMENT ON COLUMN user_verification_events.user_verification_id IS '本人確認ID';
COMMENT ON COLUMN user_verification_events.event_type IS 'イベント種別';
COMMENT ON COLUMN user_verification_events.payload IS 'イベント内容。本人確認書類画像や個人番号などの機微情報は保存しない';
COMMENT ON COLUMN user_verification_events.created_at IS '作成日時';

CREATE INDEX user_verification_events_verification_id_idx
    ON user_verification_events (user_verification_id);

-- ユーザセッション
CREATE TABLE user_sessions (
                               id uuid PRIMARY KEY,
                               user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
                               refresh_token_hash text NOT NULL,
                               ip_address inet,
                               user_agent text,
                               expires_at timestamptz NOT NULL,
                               revoked_at timestamptz,
                               created_at timestamptz NOT NULL DEFAULT NOW(),
                               updated_at timestamptz NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE user_sessions IS 'ユーザのログインセッションを管理するテーブル';
COMMENT ON COLUMN user_sessions.id IS 'セッションID。UUID v7をGo側で生成する';
COMMENT ON COLUMN user_sessions.user_id IS 'ユーザID';
COMMENT ON COLUMN user_sessions.refresh_token_hash IS 'リフレッシュトークンのハッシュ値。平文トークンは保存しない';
COMMENT ON COLUMN user_sessions.ip_address IS 'ログイン元IPアドレス';
COMMENT ON COLUMN user_sessions.user_agent IS 'ログイン元User-Agent';
COMMENT ON COLUMN user_sessions.expires_at IS 'セッション有効期限';
COMMENT ON COLUMN user_sessions.revoked_at IS 'セッション無効化日時';
COMMENT ON COLUMN user_sessions.created_at IS '作成日時';
COMMENT ON COLUMN user_sessions.updated_at IS '更新日時';

CREATE UNIQUE INDEX user_sessions_refresh_token_hash_unique_idx
    ON user_sessions (refresh_token_hash);

CREATE INDEX user_sessions_user_id_idx
    ON user_sessions (user_id);

CREATE INDEX user_sessions_active_idx
    ON user_sessions (user_id, expires_at)
    WHERE revoked_at IS NULL;

-- ログイン履歴
CREATE TABLE user_login_histories (
                                      id uuid PRIMARY KEY,
                                      user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
                                      email varchar(255),
                                      ip_address inet,
                                      user_agent text,
                                      success boolean NOT NULL,
                                      failure_reason varchar(100),
                                      created_at timestamptz NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE user_login_histories IS 'ユーザログイン履歴を管理するテーブル';
COMMENT ON COLUMN user_login_histories.id IS 'ログイン履歴ID。UUID v7をGo側で生成する';
COMMENT ON COLUMN user_login_histories.user_id IS 'ユーザID。該当ユーザが特定できない失敗ログインではNULL';
COMMENT ON COLUMN user_login_histories.email IS 'ログイン試行時に入力されたメールアドレス';
COMMENT ON COLUMN user_login_histories.ip_address IS 'ログイン元IPアドレス';
COMMENT ON COLUMN user_login_histories.user_agent IS 'ログイン元User-Agent';
COMMENT ON COLUMN user_login_histories.success IS 'ログイン成功有無';
COMMENT ON COLUMN user_login_histories.failure_reason IS 'ログイン失敗理由';
COMMENT ON COLUMN user_login_histories.created_at IS '作成日時';

CREATE INDEX user_login_histories_user_id_created_idx
    ON user_login_histories (user_id, created_at DESC);

CREATE INDEX user_login_histories_email_created_idx
    ON user_login_histories (LOWER(email), created_at DESC);

CREATE INDEX user_login_histories_ip_created_idx
    ON user_login_histories (ip_address, created_at DESC);
