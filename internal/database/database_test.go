package database

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func openTestDatabase(t *testing.T) (*Database, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "data.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	return db, path
}

func TestSetAndGet(t *testing.T) {
	db, _ := openTestDatabase(t)
	defer db.Close()

	if err := db.Set("name", "Priscilla"); err != nil {
		t.Fatal(err)
	}

	value, err := db.Get("name")
	if err != nil {
		t.Fatal(err)
	}

	if value != "Priscilla" {
		t.Fatalf("expected Priscilla, got %q", value)
	}
}

func TestGetMissingKey(t *testing.T) {
	db, _ := openTestDatabase(t)
	defer db.Close()

	_, err := db.Get("missing")

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestLastWriteWins(t *testing.T) {
	db, _ := openTestDatabase(t)
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
	db, path := openTestDatabase(t)

	if err := db.Set("city", "Dallas"); err != nil {
		t.Fatal(err)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()

	value, err := reopened.Get("city")
	if err != nil {
		t.Fatal(err)
	}

	if value != "Dallas" {
		t.Fatalf("expected Dallas, got %q", value)
	}
}

func TestDeleteSurvivesRestart(t *testing.T) {
	db, path := openTestDatabase(t)

	if err := db.Set("name", "Priscilla"); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Delete("name"); err != nil {
		t.Fatal(err)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()

	if reopened.Exists("name") {
		t.Fatal("expected deleted key to remain deleted")
	}
}

func TestIncrement(t *testing.T) {
	db, _ := openTestDatabase(t)
	defer db.Close()

	result, err := db.Increment("counter", 1)
	if err != nil {
		t.Fatal(err)
	}

	if result != 1 {
		t.Fatalf("expected 1, got %d", result)
	}
}

func TestIncrementRejectsNonInteger(t *testing.T) {
	db, _ := openTestDatabase(t)
	defer db.Close()

	if err := db.Set("counter", "hello"); err != nil {
		t.Fatal(err)
	}

	_, err := db.Increment("counter", 1)

	if !errors.Is(err, ErrNotInteger) {
		t.Fatalf("expected ErrNotInteger, got %v", err)
	}
}

func TestExpireMissingKey(t *testing.T) {
	db, _ := openTestDatabase(t)
	defer db.Close()

	updated, err := db.Expire("missing", 30)
	if err != nil {
		t.Fatal(err)
	}

	if updated {
		t.Fatal("expected missing key not to receive expiration")
	}
}

func TestTTLWithoutExpiration(t *testing.T) {
	db, _ := openTestDatabase(t)
	defer db.Close()

	if err := db.Set("name", "Priscilla"); err != nil {
		t.Fatal(err)
	}

	if ttl := db.TTL("name"); ttl != -1 {
		t.Fatalf("expected -1, got %d", ttl)
	}
}

func TestTTLMissingKey(t *testing.T) {
	db, _ := openTestDatabase(t)
	defer db.Close()

	if ttl := db.TTL("missing"); ttl != -2 {
		t.Fatalf("expected -2, got %d", ttl)
	}
}

func TestExpireAndTTL(t *testing.T) {
	db, _ := openTestDatabase(t)
	defer db.Close()

	if err := db.Set("temporary", "value"); err != nil {
		t.Fatal(err)
	}

	updated, err := db.Expire("temporary", 30)
	if err != nil {
		t.Fatal(err)
	}

	if !updated {
		t.Fatal("expected expiration to be assigned")
	}

	ttl := db.TTL("temporary")

	if ttl < 29 || ttl > 30 {
		t.Fatalf("expected TTL near 30, got %d", ttl)
	}
}

func TestExpiredKeyIsRemoved(t *testing.T) {
	db, _ := openTestDatabase(t)
	defer db.Close()

	if err := db.Set("temporary", "value"); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Expire("temporary", 0); err != nil {
		t.Fatal(err)
	}

	_, err := db.Get("temporary")

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestExpirationSurvivesRestart(t *testing.T) {
	db, path := openTestDatabase(t)

	if err := db.Set("temporary", "value"); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Expire("temporary", 2); err != nil {
		t.Fatal(err)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()

	ttl := reopened.TTL("temporary")

	if ttl < 1 || ttl > 2 {
		t.Fatalf("expected TTL near 2, got %d", ttl)
	}

	time.Sleep(3 * time.Second)

	if reopened.Exists("temporary") {
		t.Fatal("expected key to expire after restart")
	}
}

func TestSetRemovesExistingExpiration(t *testing.T) {
	db, _ := openTestDatabase(t)
	defer db.Close()

	if err := db.Set("name", "first"); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Expire("name", 30); err != nil {
		t.Fatal(err)
	}

	if err := db.Set("name", "second"); err != nil {
		t.Fatal(err)
	}

	if ttl := db.TTL("name"); ttl != -1 {
		t.Fatalf("expected -1, got %d", ttl)
	}
}

func TestFlushDB(t *testing.T) {
	db, _ := openTestDatabase(t)
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
