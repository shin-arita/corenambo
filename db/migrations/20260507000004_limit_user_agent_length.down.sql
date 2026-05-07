ALTER TABLE user_sessions
    ALTER COLUMN user_agent TYPE text;

ALTER TABLE user_login_histories
    ALTER COLUMN user_agent TYPE text;
