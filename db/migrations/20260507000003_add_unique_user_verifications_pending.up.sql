CREATE UNIQUE INDEX uq_user_verifications_pending
    ON user_verifications(user_id)
    WHERE status = 'pending';
