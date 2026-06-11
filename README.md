# Marketplace Food Price Pipeline

This project collects Indonesian marketplace prices and BMKG data as raw JSON.
Every scrape job is rate-limited, retried, logged as JSON, and recorded in
PostgreSQL.

## Setup

Requirements:

- Go 1.21 or newer
- PostgreSQL
- [Goose](https://github.com/pressly/goose)
- an Apify Tokopedia actor token for the `tokopedia` job

Create `.env` from `.env.example`, set `DATABASE_URL`, and install Goose:

```bash
go install github.com/pressly/goose/v3/cmd/goose@v3.24.1
export PATH="$(go env GOPATH)/bin:$PATH"
```

## Running Migrations

```bash
export DATABASE_URL='postgresql://user:password@localhost:5432/food_prices?sslmode=disable'
bash scripts/migrate.sh
```

Migrations live in `database/migrations` and use Goose `Up` and `Down`
sections. The legacy Python migration command remains available for local
development:

```bash
python scripts/dev.py migrate
```

## Running The Scraper

From WSL, the runner applies migrations, runs tests, then starts the scraper:

```bash
cd /mnt/d/progaming/Datathon\ Dicoding
bash ingestion/scraper-go/run_scraper.sh
```

If the Windows drive is mounted read-only for your WSL user, direct output to a
Linux-writable path:

```bash
bash ingestion/scraper-go/run_scraper.sh -output-dir /tmp/marketpulse-raw
```

Run only public BMKG jobs when no Apify token is available:

```bash
bash ingestion/scraper-go/run_scraper.sh \
  -jobs bmkg-latest-earthquake,bmkg-recent-earthquakes
```

Raw responses are written to `data/raw`. A fatal job error produces a non-zero
process exit code, allowing GitHub Actions to mark the run failed.

## Checking Run History

```sql
SELECT id, source, product_query, city, started_at, finished_at,
       records_fetched, records_failed, status, error_message
FROM scrape_runs
ORDER BY started_at DESC
LIMIT 20;
```

Each selected job creates its own row. Status changes from `running` to
`success` or `failed` even when the scrape itself cannot complete.

## Environment Variables

- `DATABASE_URL`: required PostgreSQL connection string
- `APIFY_TOKOPEDIA_TOKEN`: required for the Tokopedia job
- `SCRAPE_JOBS`: comma-separated job names
- `SCRAPE_QUERY`: marketplace query, default `beras`
- `SCRAPE_CITY`: optional run metadata, default empty
- `BMKG_ADM4`: required when the `bmkg-weather` job is selected
- `DATA_DIR`: raw JSON output directory, default `data/raw` relative to the repo

Available jobs are `tokopedia`, `bmkg-weather`, `bmkg-latest-earthquake`, and
`bmkg-recent-earthquakes`.

## Scheduling

`.github/workflows/scrape.yml` runs daily at 01:00 UTC (08:00 WIB) and can also
be started manually. Add `DATABASE_URL` and `APIFY_TOKOPEDIA_TOKEN` as GitHub
Actions repository secrets before enabling the workflow.
