package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ingestion/scraper-go/internal/scraper"
	"ingestion/scraper-go/internal/storage"
)

const defaultJobs = "tokopedia,bmkg-latest-earthquake,bmkg-recent-earthquakes"

type scrapeJob struct {
	name     string
	query    string
	city     string
	filename string
	scrape   func(context.Context) ([]byte, int, error)
}

func main() {
	os.Exit(run())
}

func run() int {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 && attr.Key == slog.TimeKey {
				attr.Key = "ts"
			}
			if len(groups) == 0 && attr.Key == slog.LevelKey {
				attr.Value = slog.StringValue(strings.ToLower(attr.Value.String()))
			}
			return attr
		},
	}))
	slog.SetDefault(logger)

	jobsValue := flag.String("jobs", envOrDefault("SCRAPE_JOBS", defaultJobs), "comma-separated scrape jobs")
	query := flag.String("query", envOrDefault("SCRAPE_QUERY", "beras"), "marketplace product query")
	city := flag.String("city", os.Getenv("SCRAPE_CITY"), "target city metadata")
	bmkgADM4 := flag.String("bmkg-adm4", os.Getenv("BMKG_ADM4"), "BMKG fourth-level administrative code")
	outputDir := flag.String("output-dir", envOrDefault("DATA_DIR", "../../data/raw"), "directory for raw JSON output")
	timeout := flag.Duration("timeout", 5*time.Minute, "maximum duration for all scrape jobs")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	store, err := storage.OpenRunStore(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("scraper initialization failed", "event", "scraper_init_failed", "reason", err)
		return 1
	}
	defer store.Close()

	jobs, err := buildJobs(*jobsValue, *query, *city, *bmkgADM4)
	if err != nil {
		logger.Error("invalid scraper configuration", "event", "scraper_init_failed", "reason", err)
		return 1
	}
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		logger.Error("create output directory", "event", "scraper_init_failed", "reason", err)
		return 1
	}

	var failures []error
	for _, job := range jobs {
		if err := executeJob(ctx, store, logger, *outputDir, job); err != nil {
			failures = append(failures, err)
		}
	}
	if len(failures) > 0 {
		logger.Error("scrape batch failed", "event", "scrape_batch_failed", "failed_jobs", len(failures))
		return 1
	}
	logger.Info("scrape batch complete", "event", "scrape_batch_complete", "jobs", len(jobs))
	return 0
}

func executeJob(ctx context.Context, store *storage.RunStore, logger *slog.Logger, outputDir string, job scrapeJob) error {
	runID, err := store.Start(ctx, storage.RunStart{Source: job.name, Query: job.query, City: job.city})
	if err != nil {
		return fmt.Errorf("%s: %w", job.name, err)
	}

	started := time.Now()
	logger.Info("scrape started",
		"event", "scrape_start",
		"run_id", runID,
		"platform", job.name,
		"query", job.query,
		"city", job.city,
	)

	data, recordsFound, scrapeErr := job.scrape(ctx)
	recordsSaved := 0
	if scrapeErr == nil {
		path := filepath.Join(outputDir, job.filename)
		if err := storage.SaveJSON(path, data); err != nil {
			scrapeErr = fmt.Errorf("save %s: %w", path, err)
		} else {
			recordsSaved = recordsFound
		}
	}

	recordsFailed := 0
	if scrapeErr != nil {
		recordsFailed = 1
	}
	finishCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	finishErr := store.Finish(finishCtx, runID, storage.RunFinish{
		RecordsFound:  recordsFound,
		RecordsSaved:  recordsSaved,
		RecordsFailed: recordsFailed,
		Err:           scrapeErr,
	})
	if finishErr != nil {
		scrapeErr = errors.Join(scrapeErr, finishErr)
	}

	duration := time.Since(started)
	if scrapeErr != nil {
		logger.Error("scrape failed",
			"event", "scrape_failed",
			"run_id", runID,
			"platform", job.name,
			"reason", scrapeErr,
			"duration_ms", duration.Milliseconds(),
		)
		return fmt.Errorf("%s: %w", job.name, scrapeErr)
	}

	logger.Info("scrape complete",
		"event", "scrape_complete",
		"run_id", runID,
		"platform", job.name,
		"fetched", recordsFound,
		"failed", recordsFailed,
		"duration_ms", duration.Milliseconds(),
	)
	return nil
}

func buildJobs(value, query, city, bmkgADM4 string) ([]scrapeJob, error) {
	available := map[string]func() scrapeJob{
		"tokopedia": func() scrapeJob {
			return scrapeJob{
				name: "tokopedia", query: query, city: city, filename: "tokopedia.json",
				scrape: func(ctx context.Context) ([]byte, int, error) {
					data, err := scraper.ScrapeTokopedia(ctx, query)
					return data, countJSONRecords(data), err
				},
			}
		},
		"bmkg-weather": func() scrapeJob {
			return scrapeJob{
				name: "bmkg-weather", query: "weather forecast", city: city, filename: "bmkg_weather_forecast.json",
				scrape: func(ctx context.Context) ([]byte, int, error) {
					if bmkgADM4 == "" {
						return nil, 0, fmt.Errorf("BMKG_ADM4 is required for bmkg-weather")
					}
					data, err := scraper.ScrapeWeatherForecast(ctx, bmkgADM4)
					return marshalResult(data, lenOrZero(data, func() int { return len(data.Data) }), err)
				},
			}
		},
		"bmkg-latest-earthquake": func() scrapeJob {
			return scrapeJob{
				name: "bmkg-latest-earthquake", query: "latest earthquake", filename: "bmkg_latest_earthquake.json",
				scrape: func(ctx context.Context) ([]byte, int, error) {
					data, err := scraper.ScrapeLatestEarthquake(ctx)
					return marshalResult(data, boolCount(data != nil), err)
				},
			}
		},
		"bmkg-recent-earthquakes": func() scrapeJob {
			return scrapeJob{
				name: "bmkg-recent-earthquakes", query: "recent earthquakes", filename: "bmkg_recent_earthquakes.json",
				scrape: func(ctx context.Context) ([]byte, int, error) {
					data, err := scraper.ScrapeRecentEarthquakes(ctx)
					return marshalResult(data, lenOrZero(data, func() int { return len(data.InfoEarthquake.Earthquakes) }), err)
				},
			}
		},
	}

	var jobs []scrapeJob
	for _, name := range strings.Split(value, ",") {
		name = strings.TrimSpace(name)
		factory, ok := available[name]
		if !ok {
			return nil, fmt.Errorf("unknown job %q", name)
		}
		jobs = append(jobs, factory())
	}
	if len(jobs) == 0 {
		return nil, fmt.Errorf("at least one scrape job is required")
	}
	return jobs, nil
}

func marshalResult(value any, count int, err error) ([]byte, int, error) {
	if err != nil {
		return nil, 0, err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, 0, fmt.Errorf("encode scrape response: %w", err)
	}
	return data, count, nil
}

func countJSONRecords(data []byte) int {
	var records []json.RawMessage
	if err := json.Unmarshal(data, &records); err == nil {
		return len(records)
	}
	if len(data) > 0 {
		return 1
	}
	return 0
}

func lenOrZero[T any](value *T, count func() int) int {
	if value == nil {
		return 0
	}
	return count()
}

func boolCount(ok bool) int {
	if ok {
		return 1
	}
	return 0
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
