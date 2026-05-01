import sys
import os

sys.path.append(os.path.dirname(os.path.dirname(__file__)))

from database.run_sql import run_sql_file

def main():
    command = sys.argv[1]

    if command == "migrate":
        run_sql_file("database/migrations/001_init.sql")

    elif command == "seed":
        run_sql_file("database/seeds/seed.sql")

    elif command == "reset":
        run_sql_file("database/migrations/001_init.sql")
        run_sql_file("database/seeds/seed.sql")

    else:
        print("Unknown command")

if __name__ == "__main__":
    main()