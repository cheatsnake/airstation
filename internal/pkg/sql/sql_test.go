package sql

import (
	dbsql "database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestBuildInClause(t *testing.T) {
	t.Run("single placeholder", func(t *testing.T) {
		got := BuildInClause("id", 1)
		want := "id IN (?)"
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("multiple placeholders", func(t *testing.T) {
		got := BuildInClause("track_id", 3)
		want := "track_id IN (?,?,?)"
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("zero produces empty IN clause", func(t *testing.T) {
		got := BuildInClause("id", 0)
		want := "id IN ()"
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})
}

func setupTestDB(t *testing.T) *dbsql.DB {
	t.Helper()
	db, err := dbsql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`CREATE TABLE test_items (id TEXT PRIMARY KEY, name TEXT)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO test_items (id, name) VALUES ('a', 'alpha')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO test_items (id, name) VALUES ('b', 'bravo')`)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func TestColumnExists(t *testing.T) {
	db := setupTestDB(t)

	t.Run("returns true for existing column", func(t *testing.T) {
		exists, err := ColumnExists(db, "test_items", "name")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Error("expected column to exist")
		}
	})

	t.Run("returns false for missing column", func(t *testing.T) {
		exists, err := ColumnExists(db, "test_items", "nonexistent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exists {
			t.Error("expected column to not exist")
		}
	})
}

func TestTableExists(t *testing.T) {
	db := setupTestDB(t)

	t.Run("returns true for existing table", func(t *testing.T) {
		exists, err := TableExists(db, "test_items")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Error("expected table to exist")
		}
	})

	t.Run("returns false for missing table", func(t *testing.T) {
		exists, err := TableExists(db, "nonexistent_table")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exists {
			t.Error("expected table to not exist")
		}
	})
}