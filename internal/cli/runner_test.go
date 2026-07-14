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

func TestSetValueContainingSpaces(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("SET greeting hello world")
	runner.Execute("GET greeting")

	expected := "OK\nhello world\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestGetMissingKey(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("GET missing")

	expected := "NULL\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestDeleteCommand(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("SET name Priscilla")
	runner.Execute("DEL name")
	runner.Execute("GET name")

	expected := "OK\n1\nNULL\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestDeleteMissingKey(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("DEL missing")

	expected := "0\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestExistsCommand(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("SET name Priscilla")
	runner.Execute("EXISTS name")
	runner.Execute("EXISTS missing")

	expected := "OK\n1\n0\n"

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

func TestMSetRejectsOddNumberOfArguments(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("MSET first 1 second")

	expected := "ERROR MSET requires key-value pairs\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestMGetRequiresKey(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("MGET")

	expected := "ERROR MGET requires at least 1 key\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestFlushDBCommand(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("SET first 1")
	runner.Execute("SET second 2")
	runner.Execute("FLUSHDB")
	runner.Execute("EXISTS first")
	runner.Execute("EXISTS second")

	expected := "OK\nOK\nOK\n0\n0\n"

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

	shouldExit := runner.Execute("EXIT")

	if !shouldExit {
		t.Fatal("expected EXIT to stop the program")
	}
}

func TestExitRejectsArguments(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	shouldExit := runner.Execute("EXIT now")

	if shouldExit {
		t.Fatal("expected invalid EXIT command not to stop the program")
	}

	expected := "ERROR EXIT does not accept arguments\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}
