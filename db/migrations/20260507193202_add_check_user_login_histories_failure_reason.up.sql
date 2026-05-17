ALTER TABLE user_login_histories
    ADD CONSTRAINT chk_user_login_histories_failure_reason
        CHECK (success = true OR failure_reason IS NOT NULL);
