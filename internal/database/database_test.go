package database

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestSetAndGet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := db.Set("name", "Priscilla"); err != nil {
		t.Fatalf("set value: %v", err)
	}

	value, err := db.Get("name")
	if err != nil {
		t.Fatalf("get value: %v", err)
	}

	if value != "Priscilla" {
		t.Fatalf("expected Priscilla, got %q", value)
	}
}

func TestGetMissingKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")

	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Get("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestLastWriteWins(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")

	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Set("name", "first"); err != nil {
		t.Fatal(err)
	}

	if err := db.Set("name", "second"); err != nil {
		t.Fatal(err)
	}

	value, err := db.Get("name")
	if err != nil {
		t.Fatal(err)
	}

	if value != "second" {
		t.Fatalf("expected second, got %q", value)
	}
}

func TestSetSurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")

	first, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := first.Set("city", "Dallas"); err != nil {
		t.Fatal(err)
	}

	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()

	value, err := second.Get("city")
	if err != nil {
		t.Fatal(err)
	}

	if value != "Dallas" {
		t.Fatalf("expected Dallas, got %q", value)
	}
}

func TestDeleteSurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")

	first, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := first.Set("name", "Priscilla"); err != nil {
		t.Fatal(err)
	}

	deleted, err := first.Delete("name")
	if err != nil {
		t.Fatal(err)
	}

	if !deleted {
		t.Fatal("expected key to be deleted")
	}

	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()

	if second.Exists("name") {
		t.Fatal("expected deleted key to remain deleted after restart")
	}
}

func TestIncrementMissingKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")

	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	result, err := db.Increment("counter", 1)
	if err != nil {
		t.Fatal(err)
	}

	if result != 1 {
		t.Fatalf("expected 1, got %d", result)
	}
}

func TestIncrementExistingKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")

	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Set("counter", "10"); err != nil {
		t.Fatal(err)
	}

	result, err := db.Increment("counter", 1)
	if err != nil {
		t.Fatal(err)
	}

	if result != 11 {
		t.Fatalf("expected 11, got %d", result)
	}
}

func TestDecrementExistingKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")

	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Set("counter", "10"); err != nil {
		t.Fatal(err)
	}

	result, err := db.Increment("counter", -1)
	if err != nil {
		t.Fatal(err)
	}

	if result != 9 {
		t.Fatalf("expected 9, got %d", result)
	}
}

func TestIncrementRejectsNonInteger(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")

	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Set("counter", "hello"); err != nil {
		t.Fatal(err)
	}

	_, err = db.Increment("counter", 1)
	if !errors.Is(err, ErrNotInteger) {
		t.Fatalf("expected ErrNotInteger, got %v", err)
	}
}

func TestIncrementSurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")

	first, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := first.Increment("counter", 1); err != nil {
		t.Fatal(err)
	}

	if _, err := first.Increment("counter", 1); err != nil {
		t.Fatal(err)
	}

	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()

	value, err := second.Get("counter")
	if err != nil {
		t.Fatal(err)
	}

	if value != "2" {
		t.Fatalf("expected 2, got %q", value)
	}
}

func TestFlushDB(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")

	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Set("first", "1"); err != nil {
		t.Fatal(err)
	}

	if err := db.Set("second", "2"); err != nil {
		t.Fatal(err)
	}

	if err := db.FlushDB(); err != nil {
		t.Fatal(err)
	}

	if db.Exists("first") || db.Exists("second") {
		t.Fatal("expected all keys to be removed")
	}
}
