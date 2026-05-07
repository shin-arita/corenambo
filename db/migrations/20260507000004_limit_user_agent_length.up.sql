ALTER TABLE user_sessions
    ALTER COLUMN user_agent TYPE varchar(1000);

ALTER TABLE user_login_histories
    ALTER COLUMN user_agent TYPE varchar(1000);
