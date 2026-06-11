-- +goose Up
CREATE TABLE IF NOT EXISTS products_raw (
    id BIGSERIAL PRIMARY KEY,
    scrape_run_id BIGINT NOT NULL REFERENCES scrape_runs(id),
    source TEXT NOT NULL CHECK (btrim(source) <> ''),
    external_product_id TEXT,
    product_title TEXT NOT NULL CHECK (btrim(product_title) <> ''),
    price_text TEXT,
    price_amount NUMERIC(18, 2) CHECK (price_amount IS NULL OR price_amount >= 0),
    currency TEXT,
    seller_name TEXT,
    seller_location_text TEXT,
    product_url TEXT,
    raw_payload JSONB NOT NULL,
    scraped_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS products_raw;
