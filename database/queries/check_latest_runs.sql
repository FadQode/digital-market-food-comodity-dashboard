SELECT
    id,
    source,
    product_query,
    city,
    status,
    started_at,
    finished_at,
    records_fetched,
    records_saved,
    records_failed,
    error_message
FROM scrape_runs
ORDER BY started_at DESC
LIMIT 20;
