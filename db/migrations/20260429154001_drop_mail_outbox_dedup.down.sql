ALTER TABLE mail_outboxes
    ADD CONSTRAINT mail_outboxes_unique_once
        UNIQUE (id, status);