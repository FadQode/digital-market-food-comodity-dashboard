package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"ingestion/scraper-go/internal/quality"
)

const insertQualityIssueSQL = `
	INSERT INTO data_quality_log (
		scrape_run_id, raw_product_id, source,
		issue_type, issue_code, severity,
		message, field_name, raw_value, details, metadata
	) VALUES (
		$1, $2, NULLIF($3, ''),
		$4, $4, $5,
		NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), $9::jsonb, $9::jsonb
	)
`

func (r *Repository) InsertQualityIssues(ctx context.Context, issues []quality.Issue) (int, error) {
	if len(issues) == 0 {
		return 0, nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin quality issue transaction: %w", err)
	}
	defer tx.Rollback()

	inserted := 0
	for _, issue := range issues {
		metadata := issue.Metadata
		if metadata == nil {
			metadata = map[string]any{}
		}
		encoded, err := json.Marshal(metadata)
		if err != nil {
			return inserted, fmt.Errorf("encode quality issue metadata: %w", err)
		}
		_, err = tx.ExecContext(ctx, insertQualityIssueSQL,
			issue.ScrapeRunID, issue.ProductRawID, issue.Source, issue.IssueCode,
			issue.Severity, issue.Message, issue.FieldName, issue.RawValue, encoded)
		if err != nil {
			return inserted, fmt.Errorf("insert quality issue %s: %w", issue.IssueCode, err)
		}
		inserted++
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit quality issues: %w", err)
	}
	return inserted, nil
}
