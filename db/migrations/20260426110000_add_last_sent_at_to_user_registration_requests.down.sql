DROP INDEX IF EXISTS idx_user_registration_requests_last_sent_at;

ALTER TABLE user_registration_requests
    DROP COLUMN IF EXISTS last_sent_at;
