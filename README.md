# Marketplace Food Price Pipeline

This project collects marketplace food-price listings, preserves the raw
responses, and prepares them for later normalization and analysis.

## Database Setup

1. Create a PostgreSQL database.
2. Copy the keys from `.env.example` into a local `.env` and set real values.
3. Install the Python dependencies:

   ```powershell
   python -m pip install -r requirements.txt
   ```

4. Apply all pending migrations:

   ```powershell
   python scripts/dev.py migrate
   ```

5. Optionally load the development seed data:

   ```powershell
   python scripts/dev.py seed
   ```

The connection can be configured with one `DATABASE_URL` value or with
`DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, and `DB_PASSWORD`. `DB_SSLMODE` is
optional and is commonly set to `require` for hosted PostgreSQL services.

## Migration Behavior

Migration files live in `database/migrations` and run in filename order. The
runner records each applied file and its SHA-256 checksum in
`schema_migrations`. Applied migration files must not be edited; add a new
numbered migration for later schema changes.

The Phase 1 data foundation contains:

- `scrape_runs`: scraper execution status, counters, timing, and errors;
- `products_raw`: original marketplace fields and the complete JSON payload;
- `data_quality_log`: visible, traceable data-quality issues;
- indexes for run history, source/timestamp lookups, URLs, raw JSON, and open
  quality issues.

Raw records reference their scrape run, and quality issues can reference a run,
a raw product, or both. The existing analytical tables remain managed by the
initial migration.

## Development Commands

```powershell
python scripts/dev.py migrate
python scripts/dev.py seed
python scripts/dev.py reset
```

`reset` is retained as a compatibility command. It applies pending migrations
and then runs the idempotent seed file; it does not drop data.
