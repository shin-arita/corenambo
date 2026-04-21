CREATE TABLE IF NOT EXISTS products (
    id            bigserial PRIMARY KEY,
    sku           varchar(64) UNIQUE NOT NULL,
    name          text NOT NULL,
    description   text NOT NULL DEFAULT '',
    price         numeric(12, 2) NOT NULL,
    is_active     boolean NOT NULL DEFAULT true,
    search_text   text GENERATED ALWAYS AS (
        coalesce(name, '') || ' ' || coalesce(description, '')
    ) STORED,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_products_search_text_pgroonga_mecab
          ON products
       USING pgroonga (search_text)
        WITH (tokenizer='TokenMecab');
