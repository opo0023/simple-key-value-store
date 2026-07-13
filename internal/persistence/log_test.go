package persistence

import (
	"path/filepath"
	"testing"
)

func TestAppendAndReplay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")

	log, err := Open(path)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}

	record := Record{
		Command: "SET",
		Args:    []string{"name", "Priscilla"},
	}

	if err := log.Append(record); err != nil {
		t.Fatalf("append record: %v", err)
	}

	if err := log.Close(); err != nil {
		t.Fatalf("close log: %v", err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen log: %v", err)
	}
	defer reopened.Close()

	count := 0

	err = reopened.Replay(func(got Record) error {
		count++

		if got.Command != "SET" {
			t.Fatalf("expected SET, got %s", got.Command)
		}

		if len(got.Args) != 2 {
			t.Fatalf("expected 2 arguments, got %d", len(got.Args))
		}

		if got.Args[0] != "name" {
			t.Fatalf("expected key name, got %q", got.Args[0])
		}

		if got.Args[1] != "Priscilla" {
			t.Fatalf("expected Priscilla, got %q", got.Args[1])
		}

		return nil
	})
	if err != nil {
		t.Fatalf("replay log: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 record, got %d", count)
	}
}

func TestTruncate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")

	log, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()

	if err := log.Append(Record{
		Command: "SET",
		Args:    []string{"key", "value"},
	}); err != nil {
		t.Fatal(err)
	}

	if err := log.Truncate(); err != nil {
		t.Fatal(err)
	}

	count := 0

	if err := log.Replay(func(record Record) error {
		count++
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Fatalf("expected empty log, got %d records", count)
	}
}
