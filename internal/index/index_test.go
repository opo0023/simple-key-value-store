package index

import (
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	store := New()

	store.Set("name", "Priscilla")

	value, exists := store.Get("name")
	if !exists {
		t.Fatal("expected key to exist")
	}

	if value != "Priscilla" {
		t.Fatalf("expected Priscilla, got %q", value)
	}
}

func TestLastWriteWins(t *testing.T) {
	store := New()

	store.Set("name", "first")
	store.Set("name", "second")

	value, exists := store.Get("name")
	if !exists {
		t.Fatal("expected key to exist")
	}

	if value != "second" {
		t.Fatalf("expected second, got %q", value)
	}
}

func TestSetRemovesPreviousExpiration(t *testing.T) {
	store := New()

	store.Set("name", "first")
	store.SetExpiration("name", time.Now().Unix()+100)
	store.Set("name", "second")

	if ttl := store.TTL("name"); ttl != -1 {
		t.Fatalf("expected TTL -1, got %d", ttl)
	}
}

func TestDelete(t *testing.T) {
	store := New()
	store.Set("name", "Priscilla")

	if !store.Delete("name") {
		t.Fatal("expected key to be deleted")
	}

	if store.Exists("name") {
		t.Fatal("expected key not to exist")
	}
}

func TestExpiration(t *testing.T) {
	store := New()
	store.Set("temporary", "value")

	if !store.SetExpiration("temporary", time.Now().Unix()-1) {
		t.Fatal("expected expiration to be assigned")
	}

	if store.Exists("temporary") {
		t.Fatal("expected expired key not to exist")
	}
}

func TestTTLWithoutExpiration(t *testing.T) {
	store := New()
	store.Set("name", "Priscilla")

	if ttl := store.TTL("name"); ttl != -1 {
		t.Fatalf("expected -1, got %d", ttl)
	}
}

func TestTTLMissingKey(t *testing.T) {
	store := New()

	if ttl := store.TTL("missing"); ttl != -2 {
		t.Fatalf("expected -2, got %d", ttl)
	}
}

func TestTTLWithExpiration(t *testing.T) {
	store := New()
	store.Set("temporary", "value")

	expiresAt := time.Now().Unix() + 30

	if !store.SetExpiration("temporary", expiresAt) {
		t.Fatal("expected expiration to be assigned")
	}

	ttl := store.TTL("temporary")

	if ttl < 29 || ttl > 30 {
		t.Fatalf("expected TTL near 30, got %d", ttl)
	}
}

func TestClear(t *testing.T) {
	store := New()
	store.Set("first", "1")
	store.Set("second", "2")

	store.Clear()

	if store.Exists("first") || store.Exists("second") {
		t.Fatal("expected all keys to be removed")
	}
}
