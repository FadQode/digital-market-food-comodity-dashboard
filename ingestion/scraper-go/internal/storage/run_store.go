package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
)

type RunStore struct {
	db *sql.DB
}

type RunStart struct {
	Source string
	Query  string
	City   string
}

type RunFinish struct {
	RecordsFound  int
	RecordsSaved  int
	RecordsFailed int
	Err           error
}

func OpenRunStore(ctx context.Context, databaseURL string) (*RunStore, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	return &RunStore{db: db}, nil
}

func (s *RunStore) Close() error {
	return s.db.Close()
}

func (s *RunStore) Start(ctx context.Context, run RunStart) (int64, error) {
	metadata, err := json.Marshal(map[string]string{
		"product_query": run.Query,
		"city":          run.City,
	})
	if err != nil {
		return 0, fmt.Errorf("encode run metadata: %w", err)
	}

	var id int64
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO scrape_runs (source, product_query, city, status, metadata)
		VALUES ($1, $2, NULLIF($3, ''), 'running', $4::jsonb)
		RETURNING id
	`, run.Source, run.Query, run.City, metadata).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("start scrape run: %w", err)
	}
	return id, nil
}

func (s *RunStore) Finish(ctx context.Context, id int64, result RunFinish) error {
	status := "success"
	errorMessage := ""
	if result.Err != nil {
		status = "failed"
		errorMessage = result.Err.Error()
	}

	resultValue, err := s.db.ExecContext(ctx, `
		UPDATE scrape_runs
		SET finished_at = CURRENT_TIMESTAMP,
			status = $2,
			records_found = $3,
			records_saved = $4,
			records_failed = $5,
			records_fetched = $3,
			error_message = NULLIF($6, '')
		WHERE id = $1
	`, id, status, result.RecordsFound, result.RecordsSaved, result.RecordsFailed, errorMessage)
	if err != nil {
		return fmt.Errorf("finish scrape run %d: %w", id, err)
	}
	rows, err := resultValue.RowsAffected()
	if err != nil {
		return fmt.Errorf("read scrape run update result: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("finish scrape run %d: row not found", id)
	}
	return nil
}
