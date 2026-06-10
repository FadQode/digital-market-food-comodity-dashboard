import hashlib
import os
from pathlib import Path

import psycopg2
from dotenv import load_dotenv


PROJECT_ROOT = Path(__file__).resolve().parents[1]
MIGRATIONS_DIR = PROJECT_ROOT / "database" / "migrations"

load_dotenv(PROJECT_ROOT / ".env")


def connect_database():
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
                if previous_checksum.strip() != checksum:
                    raise RuntimeError(
                        f"Applied migration has changed: {path.name}. "
                        "Create a new migration instead of editing it."
                    )
                continue

            with conn:
                with conn.cursor() as cursor:
                    cursor.execute(path.read_text(encoding="utf-8"))
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
