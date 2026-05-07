ALTER TABLE user_emails
    ADD CONSTRAINT chk_user_emails_primary_requires_verified
        CHECK (is_primary = false OR verified_at IS NOT NULL);
