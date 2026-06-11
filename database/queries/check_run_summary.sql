SELECT
    sr.id,
    sr.source,
    sr.product_query,
    sr.status,
    sr.started_at,
    sr.finished_at,
    sr.records_fetched,
    sr.records_saved,
    sr.records_failed,
    COALESCE(pr.products_raw_count, 0) AS products_raw_count,
    COALESCE(dq.quality_issue_count, 0) AS quality_issue_count
FROM scrape_runs sr
LEFT JOIN (
    SELECT scrape_run_id, COUNT(*) AS products_raw_count
    FROM products_raw
    GROUP BY scrape_run_id
) pr ON pr.scrape_run_id = sr.id
LEFT JOIN (
    SELECT scrape_run_id, COUNT(*) AS quality_issue_count
    FROM data_quality_log
    GROUP BY scrape_run_id
) dq ON dq.scrape_run_id = sr.id
ORDER BY sr.started_at DESC
LIMIT 20;
