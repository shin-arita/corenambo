CREATE INDEX idx_items_search
ON items
USING pgroonga (
    title,
    description
);
