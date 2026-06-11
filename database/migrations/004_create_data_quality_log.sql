-- +goose Up
CREATE TABLE IF NOT EXISTS data_quality_log (
    id BIGSERIAL PRIMARY KEY,
    scrape_run_id BIGINT REFERENCES scrape_runs(id),
    raw_product_id BIGINT REFERENCES products_raw(id),
    issue_type TEXT NOT NULL CHECK (btrim(issue_type) <> ''),
    severity TEXT NOT NULL DEFAULT 'warning'
        CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    message TEXT,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    resolved_at TIMESTAMPTZ,
    resolution_note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (scrape_run_id IS NOT NULL OR raw_product_id IS NOT NULL),
    CHECK (resolved_at IS NULL OR resolved_at >= created_at)
);

-- +goose Down
DROP TABLE IF EXISTS data_quality_log;
