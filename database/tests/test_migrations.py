import tempfile
import unittest
from pathlib import Path

from database.run_sql import discover_migrations, run_migrations, run_sql_file


class FakeCursor:
    def __init__(self, connection):
        self.connection = connection
        self.rows = []

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        return False

    def execute(self, sql, parameters=None):
        normalized = " ".join(sql.split())
        self.connection.statements.append((normalized, parameters))

        if normalized.startswith("SELECT filename, checksum"):
            self.rows = list(self.connection.applied.items())
        elif normalized.startswith("INSERT INTO schema_migrations"):
            filename, checksum = parameters
            self.connection.applied[filename] = checksum

    def fetchall(self):
        return self.rows


class FakeConnection:
    def __init__(self):
        self.applied = {}
        self.statements = []
        self.closed = False

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        return False

    def cursor(self):
        return FakeCursor(self)

    def close(self):
        self.closed = True


class MigrationTests(unittest.TestCase):
    def test_discovers_numbered_migrations_in_order(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "010_later.sql").write_text("SELECT 10;", encoding="utf-8")
            (root / "002_earlier.sql").write_text("SELECT 2;", encoding="utf-8")
            (root / "notes.sql").write_text("SELECT 0;", encoding="utf-8")

            names = [path.name for path in discover_migrations(root)]

        self.assertEqual(names, ["002_earlier.sql", "010_later.sql"])

    def test_applies_each_migration_once(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "001_first.sql").write_text("SELECT 1;", encoding="utf-8")
            (root / "002_second.sql").write_text("SELECT 2;", encoding="utf-8")
            connection = FakeConnection()

            self.assertEqual(run_migrations(root, connection), 2)
            self.assertEqual(run_migrations(root, connection), 0)

        self.assertEqual(
            list(connection.applied), ["001_first.sql", "002_second.sql"]
        )
        self.assertFalse(connection.closed)

    def test_rejects_changed_applied_migration(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            migration = root / "001_first.sql"
            migration.write_text("SELECT 1;", encoding="utf-8")
            connection = FakeConnection()
            run_migrations(root, connection)
            migration.write_text("SELECT 2;", encoding="utf-8")

            with self.assertRaisesRegex(RuntimeError, "Applied migration has changed"):
                run_migrations(root, connection)

    def test_executes_standalone_sql_file(self):
        with tempfile.TemporaryDirectory() as directory:
            sql_file = Path(directory) / "seed.sql"
            sql_file.write_text("SELECT 42;", encoding="utf-8")
            connection = FakeConnection()

            run_sql_file(sql_file, connection)

        self.assertTrue(
            any(statement == "SELECT 42;" for statement, _ in connection.statements)
        )


if __name__ == "__main__":
    unittest.main()
