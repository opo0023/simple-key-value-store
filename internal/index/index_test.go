package index

import "testing"

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

func TestGetMissingKey(t *testing.T) {
	store := New()

	_, exists := store.Get("missing")
	if exists {
		t.Fatal("expected missing key not to exist")
	}
}

func TestDelete(t *testing.T) {
	store := New()
	store.Set("name", "Priscilla")

	deleted := store.Delete("name")
	if !deleted {
		t.Fatal("expected key to be deleted")
	}

	if store.Exists("name") {
		t.Fatal("expected key not to exist after deletion")
	}
}

func TestDeleteMissingKey(t *testing.T) {
	store := New()

	deleted := store.Delete("missing")
	if deleted {
		t.Fatal("expected deleting a missing key to return false")
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
