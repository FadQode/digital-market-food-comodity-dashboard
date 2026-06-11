-- +goose Up
ALTER TABLE products_raw
    DROP CONSTRAINT IF EXISTS products_raw_product_title_check;

ALTER TABLE products_raw
    ALTER COLUMN product_title DROP NOT NULL,
    ADD COLUMN IF NOT EXISTS source_product_id TEXT,
    ADD COLUMN IF NOT EXISTS source_shop_id TEXT,
    ADD COLUMN IF NOT EXISTS image_url TEXT,
    ADD COLUMN IF NOT EXISTS product_name TEXT,
    ADD COLUMN IF NOT EXISTS shop_name TEXT,
    ADD COLUMN IF NOT EXISTS price_value NUMERIC(18, 2),
    ADD COLUMN IF NOT EXISTS rating NUMERIC(4, 2),
    ADD COLUMN IF NOT EXISTS review_count INTEGER,
    ADD COLUMN IF NOT EXISTS sold_count INTEGER,
    ADD COLUMN IF NOT EXISTS stock_text TEXT,
    ADD COLUMN IF NOT EXISTS result_rank INTEGER,
    ADD COLUMN IF NOT EXISTS source_query TEXT;

UPDATE products_raw
SET source_product_id = COALESCE(source_product_id, external_product_id),
    product_name = COALESCE(product_name, product_title),
    shop_name = COALESCE(shop_name, seller_name),
    price_value = COALESCE(price_value, price_amount),
    currency = COALESCE(NULLIF(currency, ''), 'IDR');

ALTER TABLE products_raw
    ALTER COLUMN currency SET DEFAULT 'IDR';

ALTER TABLE data_quality_log
    ADD COLUMN IF NOT EXISTS source TEXT,
    ADD COLUMN IF NOT EXISTS issue_code TEXT,
    ADD COLUMN IF NOT EXISTS field_name TEXT,
    ADD COLUMN IF NOT EXISTS raw_value TEXT,
    ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}'::jsonb;

UPDATE data_quality_log
SET issue_code = COALESCE(issue_code, issue_type),
    metadata = CASE
        WHEN metadata = '{}'::jsonb THEN details
        ELSE metadata
    END;

ALTER TABLE data_quality_log
    ALTER COLUMN issue_code SET NOT NULL;

ALTER TABLE data_quality_log
    DROP CONSTRAINT IF EXISTS data_quality_log_issue_code_check;

ALTER TABLE data_quality_log
    ADD CONSTRAINT data_quality_log_issue_code_check
    CHECK (btrim(issue_code) <> '');

DROP INDEX IF EXISTS idx_products_raw_run_external_id;

CREATE UNIQUE INDEX IF NOT EXISTS products_raw_run_source_product_uidx
    ON products_raw (scrape_run_id, source, source_product_id)
    WHERE source_product_id IS NOT NULL AND btrim(source_product_id) <> '';

CREATE UNIQUE INDEX IF NOT EXISTS products_raw_run_source_url_uidx
    ON products_raw (scrape_run_id, source, product_url)
    WHERE product_url IS NOT NULL AND btrim(product_url) <> '';

CREATE INDEX IF NOT EXISTS idx_products_raw_source_query_scraped_at
    ON products_raw (source, source_query, scraped_at DESC);

CREATE INDEX IF NOT EXISTS idx_data_quality_log_issue_code_created_at
    ON data_quality_log (issue_code, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_data_quality_log_issue_code_created_at;
DROP INDEX IF EXISTS idx_products_raw_source_query_scraped_at;
DROP INDEX IF EXISTS products_raw_run_source_url_uidx;
DROP INDEX IF EXISTS products_raw_run_source_product_uidx;

UPDATE data_quality_log
SET issue_type = COALESCE(NULLIF(issue_type, ''), issue_code),
    details = CASE
        WHEN details = '{}'::jsonb THEN metadata
        ELSE details
    END;

ALTER TABLE data_quality_log
    DROP CONSTRAINT IF EXISTS data_quality_log_issue_code_check,
    DROP COLUMN IF EXISTS metadata,
    DROP COLUMN IF EXISTS raw_value,
    DROP COLUMN IF EXISTS field_name,
    DROP COLUMN IF EXISTS issue_code,
    DROP COLUMN IF EXISTS source;

UPDATE products_raw
SET external_product_id = COALESCE(external_product_id, source_product_id),
    product_title = COALESCE(NULLIF(product_title, ''), NULLIF(product_name, ''), '[missing]'),
    seller_name = COALESCE(seller_name, shop_name),
    price_amount = COALESCE(price_amount, price_value);

CREATE UNIQUE INDEX IF NOT EXISTS idx_products_raw_run_external_id
    ON products_raw (scrape_run_id, source, external_product_id)
    WHERE external_product_id IS NOT NULL;

ALTER TABLE products_raw
    DROP COLUMN IF EXISTS source_query,
    DROP COLUMN IF EXISTS result_rank,
    DROP COLUMN IF EXISTS stock_text,
    DROP COLUMN IF EXISTS sold_count,
    DROP COLUMN IF EXISTS review_count,
    DROP COLUMN IF EXISTS rating,
    DROP COLUMN IF EXISTS price_value,
    DROP COLUMN IF EXISTS shop_name,
    DROP COLUMN IF EXISTS product_name,
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS source_shop_id,
    DROP COLUMN IF EXISTS source_product_id;

ALTER TABLE products_raw
    ALTER COLUMN currency DROP DEFAULT,
    ALTER COLUMN product_title SET NOT NULL;

ALTER TABLE products_raw
    ADD CONSTRAINT products_raw_product_title_check
    CHECK (btrim(product_title) <> '');
