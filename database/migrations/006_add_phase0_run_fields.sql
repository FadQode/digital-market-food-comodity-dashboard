-- +goose Up
ALTER TABLE scrape_runs
    ADD COLUMN IF NOT EXISTS product_query TEXT,
    ADD COLUMN IF NOT EXISTS city TEXT,
    ADD COLUMN IF NOT EXISTS records_fetched INTEGER NOT NULL DEFAULT 0
        CHECK (records_fetched >= 0);

UPDATE scrape_runs
SET product_query = COALESCE(NULLIF(metadata ->> 'product_query', ''), 'unspecified'),
    city = COALESCE(city, NULLIF(metadata ->> 'city', '')),
    records_fetched = records_found
WHERE product_query IS NULL;

ALTER TABLE scrape_runs
    ALTER COLUMN product_query SET NOT NULL;

ALTER TABLE scrape_runs
    DROP CONSTRAINT IF EXISTS scrape_runs_status_check;

ALTER TABLE scrape_runs
    ADD CONSTRAINT scrape_runs_status_check
    CHECK (status IN ('running', 'success', 'succeeded', 'partial', 'failed'));

-- +goose Down
UPDATE scrape_runs SET status = 'succeeded' WHERE status = 'success';

ALTER TABLE scrape_runs
    DROP CONSTRAINT IF EXISTS scrape_runs_status_check;

ALTER TABLE scrape_runs
    ADD CONSTRAINT scrape_runs_status_check
    CHECK (status IN ('running', 'succeeded', 'partial', 'failed'));

ALTER TABLE scrape_runs
    DROP COLUMN IF EXISTS records_fetched,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS product_query;
