-- +goose Up
CREATE INDEX IF NOT EXISTS idx_scrape_runs_source_started_at
    ON scrape_runs (source, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_scrape_runs_status_started_at
    ON scrape_runs (status, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_products_raw_scrape_run_id
    ON products_raw (scrape_run_id);

CREATE INDEX IF NOT EXISTS idx_products_raw_source_scraped_at
    ON products_raw (source, scraped_at DESC);

CREATE INDEX IF NOT EXISTS idx_products_raw_product_url
    ON products_raw (product_url)
    WHERE product_url IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_products_raw_run_external_id
    ON products_raw (scrape_run_id, source, external_product_id)
    WHERE external_product_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_products_raw_payload_gin
    ON products_raw USING GIN (raw_payload);

CREATE INDEX IF NOT EXISTS idx_data_quality_log_raw_product_id
    ON data_quality_log (raw_product_id)
    WHERE raw_product_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_data_quality_log_run_created_at
    ON data_quality_log (scrape_run_id, created_at DESC)
    WHERE scrape_run_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_data_quality_log_open_issues
    ON data_quality_log (severity, issue_type, created_at DESC)
    WHERE resolved_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_data_quality_log_open_issues;
DROP INDEX IF EXISTS idx_data_quality_log_run_created_at;
DROP INDEX IF EXISTS idx_data_quality_log_raw_product_id;
DROP INDEX IF EXISTS idx_products_raw_payload_gin;
DROP INDEX IF EXISTS idx_products_raw_run_external_id;
DROP INDEX IF EXISTS idx_products_raw_product_url;
DROP INDEX IF EXISTS idx_products_raw_source_scraped_at;
DROP INDEX IF EXISTS idx_products_raw_scrape_run_id;
DROP INDEX IF EXISTS idx_scrape_runs_status_started_at;
DROP INDEX IF EXISTS idx_scrape_runs_source_started_at;
