CREATE UNIQUE INDEX uq_user_registration_requests_email
    ON user_registration_requests (LOWER(email));
