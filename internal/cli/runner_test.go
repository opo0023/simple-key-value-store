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

func TestIncrementCommand(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("INCR counter")
	runner.Execute("INCR counter")
	runner.Execute("GET counter")

	expected := "1\n2\n2\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestDecrementCommand(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("SET counter 10")
	runner.Execute("DECR counter")
	runner.Execute("GET counter")

	expected := "OK\n9\n9\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestDecrementMissingKey(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("DECR counter")

	expected := "-1\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestIncrementRejectsNonInteger(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("SET counter hello")
	runner.Execute("INCR counter")

	expected := "OK\nERROR value is not an integer\n"

	if output.String() != expected {
		t.Fatalf("expected %q, got %q", expected, output.String())
	}
}

func TestIncrementRejectsMissingArgument(t *testing.T) {
	runner, output, db := newTestRunner(t)
	defer db.Close()

	runner.Execute("INCR")

	expected := "ERROR counter command requires 1 argument\n"

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
