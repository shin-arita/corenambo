CREATE INDEX idx_mail_outboxes_pending
ON mail_outboxes (status, next_attempt_at);
