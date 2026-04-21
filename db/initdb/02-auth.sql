CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    email varchar(255) UNIQUE NOT NULL,
    password_hash text NOT NULL,
    name varchar(255) NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_refresh_tokens (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash text NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_refresh_tokens_user_id
    ON user_refresh_tokens (user_id);

CREATE INDEX IF NOT EXISTS idx_user_refresh_tokens_token_hash
    ON user_refresh_tokens (token_hash);

INSERT INTO users (email, password_hash, name)
VALUES (
    'test@example.com',
    crypt('password123', gen_salt('bf')),
    'Test User'
)
ON CONFLICT (email) DO NOTHING;
