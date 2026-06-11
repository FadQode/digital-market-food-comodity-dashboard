# Marketplace Food Price Pipeline

This project collects Indonesian marketplace prices and BMKG data as raw JSON.
Every scrape job is logged as JSON and recorded in PostgreSQL. Public BMKG GET
requests use rate limiting and retries; paid marketplace Actor POST requests
use hard cost limits and are never retried automatically.

## Setup

Requirements:

- Go 1.21 or newer
- PostgreSQL
- [Goose](https://github.com/pressly/goose)
- Apify tokens for paid Tokopedia and Shopee jobs

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
sections. Goose is the canonical migration tool used by local runs and GitHub
Actions. The legacy Python command remains available for existing local
databases, but do not alternate migration tools for the same database because
they maintain separate migration histories:

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

Paid marketplace jobs are blocked unless explicitly enabled. A conservative
pilot run collects at most 20 results per marketplace and caps each Actor run
at USD 0.07:

```bash
bash ingestion/scraper-go/run_scraper.sh \
  -jobs tokopedia,shopee \
  -query "beras 5 kg" \
  -marketplace-max-items 20 \
  -apify-max-charge-usd 0.07 \
  -print-plan
```

`-print-plan` does not connect to PostgreSQL or external APIs and does not print
tokens. After reviewing that output, enable the paid calls explicitly:

```bash
bash ingestion/scraper-go/run_scraper.sh \
  -jobs tokopedia,shopee \
  -query "beras 5 kg" \
  -allow-paid-apis \
  -marketplace-max-items 20 \
  -apify-max-charge-usd 0.07
```

The Apify Actor POST is intentionally not retried because a retry can start a
second paid run when the first response is lost. HTTP retry/backoff remains
enabled for public BMKG GET requests.

Raw responses are written to `data/raw` with names such as
`run-42-tokopedia.json`, so a later scrape does not overwrite an earlier run.
A fatal job error produces a non-zero process exit code, allowing GitHub
Actions to mark the run failed.

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

For marketplace jobs, counters have these meanings:

- `records_fetched`: listing items decoded from the Actor payload
- `records_saved`: rows inserted into `products_raw`
- `records_failed`: listing rows rejected by storage or structurally unusable

## Raw Data Storage

Tokopedia and Shopee jobs insert each decoded listing into `products_raw` after
the complete Actor response has been saved to disk. Every row links back to its
`scrape_runs.id`, exposes common inspection fields, and preserves the complete
marketplace item in `raw_payload`.

```sql
SELECT id, scrape_run_id, source, product_name, price_text, price_value,
       seller_location_text, result_rank
FROM products_raw
ORDER BY created_at DESC
LIMIT 20;
```

`price_value` is the raw listing price. It is not yet normalized by package
weight, variant, voucher, bundle, or minimum order, so it must not be presented
as a commodity price per kilogram.

## Data Quality Issues

Phase 1 flags missing names, prices, URLs, and seller locations; duplicate
products within a run; suspicious raw prices; empty Actor results; and adapter
decode failures. These issues do not clean or silently discard the source data.

```sql
SELECT severity, issue_code, COUNT(*)
FROM data_quality_log
GROUP BY severity, issue_code
ORDER BY severity, issue_code;
```

Reusable inspection queries are available in `database/queries`:

- `check_latest_runs.sql`
- `check_products_raw.sql`
- `check_quality_issues.sql`
- `check_run_summary.sql`

## Environment Variables

- `DATABASE_URL`: required PostgreSQL connection string
- `TOKOPEDIA_APIFY_KEY`: required for the Tokopedia job
- `SHOPEE_APIFY_KEY`: required for the Shopee job
- `TOKOPEDIA_APIFY_ACTOR`: optional Actor override
- `SHOPEE_APIFY_ACTOR`: optional Actor override
- `SCRAPE_JOBS`: comma-separated job names
- `SCRAPE_QUERY`: marketplace query, default `beras 5 kg`
- `SCRAPE_CITY`: optional run metadata, default empty
- `ALLOW_PAID_APIS`: must be `true` before paid jobs can run
- `MARKETPLACE_MAX_ITEMS`: paid results per marketplace job, default `20`
- `APIFY_MAX_CHARGE_USD`: hard charge cap per Actor invocation, default `0.07`
- `BMKG_ADM4`: required when the `bmkg-weather` job is selected
- `DATA_DIR`: raw JSON output directory, default `data/raw` relative to the repo

Available jobs are `tokopedia`, `shopee`, `bmkg-weather`,
`bmkg-latest-earthquake`, and `bmkg-recent-earthquakes`.

`SCRAPE_CITY` is run metadata; it does not currently filter Actor search
results. Seller location must be normalized from each returned listing.

## Scheduling

`.github/workflows/scrape.yml` runs each Monday at 01:00 UTC (08:00 WIB) and can
also be started manually. Add `DATABASE_URL`, `TOKOPEDIA_APIFY_KEY`, and
`SHOPEE_APIFY_KEY` as GitHub Actions repository secrets. Each workflow run
uploads its raw JSON files as a GitHub Actions artifact retained for 14 days.

The sampling and budget rationale is documented in
[`docs/MINING_STRATEGY.md`](docs/MINING_STRATEGY.md).
