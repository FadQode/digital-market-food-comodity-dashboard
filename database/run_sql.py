import hashlib
import os
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parents[1]
MIGRATIONS_DIR = PROJECT_ROOT / "database" / "migrations"
LEGACY_MIGRATION_CHECKSUMS = {
    "001_init.sql": "e811230cccaa001d478bb1ba06b18d1022f501a0bf1bd786a08b1677585df107",
    "002_create_scrape_runs.sql": "689a239b3965343a18b9c26b46e75e3f0c1c9b00096b70737d5e1ed6d06863d5",
    "003_create_products_raw.sql": "f0479c342ec6e21ea184db898bad0a1610b4f38135108d657d634f3dc7f7e24f",
    "004_create_data_quality_log.sql": "8e424291d2392c40616a7cbff9abfe9dee910937e9dbae4275546a9660a0e283",
    "005_add_data_foundation_indexes.sql": "d9d1f7494a198ee8f05e31f06ff9a65523490b71ec112ab30c9854f01050592d",
}

try:
    from dotenv import load_dotenv
except ImportError:
    load_dotenv = None

if load_dotenv is not None:
    load_dotenv(PROJECT_ROOT / ".env")


def connect_database():
    try:
        import psycopg2
    except ImportError as exc:
        raise RuntimeError(
            "psycopg2 is required for database commands; install requirements.txt"
        ) from exc

    database_url = os.getenv("DATABASE_URL")
    if database_url:
        return psycopg2.connect(database_url)

    env_to_parameter = {
        "DB_HOST": "host",
        "DB_PORT": "port",
        "DB_NAME": "dbname",
        "DB_USER": "user",
        "DB_PASSWORD": "password",
    }
    missing = [name for name in env_to_parameter if not os.getenv(name)]
    if missing:
        names = ", ".join(missing)
        raise RuntimeError(
            f"Database configuration is incomplete. Set DATABASE_URL or: {names}"
        )

    parameters = {
        parameter: os.environ[env_name]
        for env_name, parameter in env_to_parameter.items()
    }
    if os.getenv("DB_SSLMODE"):
        parameters["sslmode"] = os.environ["DB_SSLMODE"]

    return psycopg2.connect(**parameters)


def discover_migrations(migrations_dir=MIGRATIONS_DIR):
    return sorted(Path(migrations_dir).glob("[0-9]*_*.sql"))


def _checksum(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def _read_up_migration(path):
    sql = path.read_text(encoding="utf-8")
    up_marker = "-- +goose Up"
    down_marker = "-- +goose Down"
    if up_marker not in sql:
        return sql
    up_sql = sql.split(up_marker, 1)[1]
    return up_sql.split(down_marker, 1)[0]


def run_migrations(migrations_dir=MIGRATIONS_DIR, connection=None):
    migrations = discover_migrations(migrations_dir)
    if not migrations:
        raise RuntimeError(f"No migrations found in {migrations_dir}")

    owns_connection = connection is None
    conn = connection or connect_database()

    try:
        with conn:
            with conn.cursor() as cursor:
                cursor.execute(
                    """
                    CREATE TABLE IF NOT EXISTS schema_migrations (
                        filename TEXT PRIMARY KEY,
                        checksum CHAR(64) NOT NULL,
                        applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
                    )
                    """
                )
                cursor.execute("SELECT filename, checksum FROM schema_migrations")
                applied = dict(cursor.fetchall())

        applied_count = 0
        for path in migrations:
            checksum = _checksum(path)
            previous_checksum = applied.get(path.name)

            if previous_checksum:
                previous_checksum = previous_checksum.strip()
                if previous_checksum == checksum:
                    continue
                if previous_checksum == LEGACY_MIGRATION_CHECKSUMS.get(path.name):
                    with conn:
                        with conn.cursor() as cursor:
                            cursor.execute(
                                "UPDATE schema_migrations SET checksum = %s "
                                "WHERE filename = %s",
                                (checksum, path.name),
                            )
                    continue
                raise RuntimeError(
                    f"Applied migration has changed: {path.name}. "
                    "Create a new migration instead of editing it."
                )

            with conn:
                with conn.cursor() as cursor:
                    cursor.execute(_read_up_migration(path))
                    cursor.execute(
                        "INSERT INTO schema_migrations (filename, checksum) VALUES (%s, %s)",
                        (path.name, checksum),
                    )

            applied_count += 1
            print(f"Applied migration: {path.name}")

        if applied_count == 0:
            print("Database schema is up to date.")
        return applied_count
    finally:
        if owns_connection:
            conn.close()


def run_sql_file(path, connection=None):
    sql_path = Path(path)
    owns_connection = connection is None
    conn = connection or connect_database()

    try:
        with conn:
            with conn.cursor() as cursor:
                cursor.execute(sql_path.read_text(encoding="utf-8"))
        print(f"Executed SQL file: {sql_path}")
    finally:
        if owns_connection:
            conn.close()
