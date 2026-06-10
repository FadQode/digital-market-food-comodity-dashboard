import argparse
import sys
from pathlib import Path


PROJECT_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(PROJECT_ROOT))

from database.run_sql import run_migrations, run_sql_file


def parse_args():
    parser = argparse.ArgumentParser(description="Local database development commands")
    parser.add_argument("command", choices=("migrate", "seed", "reset"))
    return parser.parse_args()


def main():
    args = parse_args()

    if args.command == "migrate":
        run_migrations()
        return

    if args.command == "reset":
        # Kept for compatibility: this applies pending migrations before seeding.
        run_migrations()

    run_sql_file(PROJECT_ROOT / "database" / "seeds" / "seed.sql")


if __name__ == "__main__":
    main()
