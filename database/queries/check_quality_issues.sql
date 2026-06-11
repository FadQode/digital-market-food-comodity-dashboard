SELECT
    severity,
    issue_code,
    COUNT(*) AS issue_count
FROM data_quality_log
GROUP BY severity, issue_code
ORDER BY severity, issue_code;
