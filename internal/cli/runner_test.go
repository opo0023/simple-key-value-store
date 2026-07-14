package cli

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/opo0023/simple-key-value-store/internal/database"
)

func newTestRunner(t *testing.T) (*Runner, *bytes.Buffer, *database.Database) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "data.db")

	db, err := database.Open(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	output := &bytes.Buffer{}
	runner := NewRunner(db, output)

	return runner, output, db
}

func TestSetAndGetCommands(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("SET name Priscilla")
	runner.Execute("GET name")

	expected := "OK\nPriscilla\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestDeleteAndExistsCommands(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("SET name Priscilla")
	runner.Execute("EXISTS name")
	runner.Execute("DEL name")
	runner.Execute("EXISTS name")

	expected := "OK\n1\n1\n0\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestMSetAndMGetCommands(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("MSET first 1 second 2 third 3")
	runner.Execute("MGET first second missing third")

	expected := "OK\n1\n2\nNULL\n3\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestIncrementAndDecrementCommands(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("INCR counter")
	runner.Execute("INCR counter")
	runner.Execute("DECR counter")

	expected := "1\n2\n1\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestExpireAndTTLCommands(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("SET session active")
	runner.Execute("EXPIRE session 30")
	runner.Execute("TTL session")

	lines := output.String()

	if lines != "OK\n1\n30\n" && lines != "OK\n1\n29\n" {
		t.Fatalf("unexpected output %q", lines)
	}
}

func TestExpireMissingKey(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("EXPIRE missing 30")

	expected := "0\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestTTLMissingKey(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("TTL missing")

	expected := "-2\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestTTLWithoutExpiration(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("SET name Priscilla")
	runner.Execute("TTL name")

	expected := "OK\n-1\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestExpireRejectsInvalidSeconds(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("SET name Priscilla")
	runner.Execute("EXPIRE name invalid")
	runner.Execute("EXPIRE name -10")

	expected := "OK\nERROR expiration must be a nonnegative integer\n" +
		"ERROR expiration must be a nonnegative integer\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestFlushDBCommand(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("SET first 1")
	runner.Execute("FLUSHDB")
	runner.Execute("EXISTS first")

	expected := "OK\nOK\n0\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestUnknownCommand(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("INVALID")

	expected := "ERROR unknown command\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestExitCommand(t *testing.T) {
	runner, _, db := newTestRunner(t)
	defer db.Close()

	if !runner.Execute("EXIT") {
		t.Fatal("expected EXIT to stop the program")
	}
}
