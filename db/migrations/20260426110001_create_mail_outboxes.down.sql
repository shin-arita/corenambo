DROP INDEX IF EXISTS idx_mail_outboxes_created_at;
DROP INDEX IF EXISTS idx_mail_outboxes_status_next_attempt_at;

DROP TABLE IF EXISTS mail_outboxes;
