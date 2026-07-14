package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/opo0023/simple-key-value-store/internal/database"
)

// Runner reads commands and writes command results.
type Runner struct {
	db  *database.Database
	out io.Writer
}

// NewRunner creates a new CLI command runner.
func NewRunner(db *database.Database, out io.Writer) *Runner {
	return &Runner{
		db:  db,
		out: out,
	}
}

// Execute processes one command.
//
// It returns true when the program should exit.
func (r *Runner) Execute(line string) bool {
	line = strings.TrimSpace(line)

	if line == "" {
		return false
	}

	parts := strings.Fields(line)
	command := strings.ToUpper(parts[0])

	switch command {
	case "SET":
		r.executeSet(line)

	case "GET":
		r.executeGet(parts)

	case "DEL":
		r.executeDelete(parts)

	case "EXISTS":
		r.executeExists(parts)

	case "MSET":
		r.executeMSet(parts)

	case "MGET":
		r.executeMGet(parts)

	case "FLUSHDB":
		r.executeFlushDB(parts)

	case "EXIT":
		if len(parts) != 1 {
			r.printError("EXIT does not accept arguments")
			return false
		}

		return true

	default:
		r.printError("unknown command")
	}

	return false
}

func (r *Runner) executeSet(line string) {
	parts := strings.SplitN(line, " ", 3)

	if len(parts) != 3 {
		r.printError("SET requires a key and value")
		return
	}

	key := strings.TrimSpace(parts[1])
	value := parts[2]

	if key == "" {
		r.printError("SET requires a key and value")
		return
	}

	if err := r.db.Set(key, value); err != nil {
		r.printError(err.Error())
		return
	}

	fmt.Fprintln(r.out, "OK")
}

func (r *Runner) executeGet(parts []string) {
	if len(parts) != 2 {
		r.printError("GET requires 1 argument")
		return
	}

	value, err := r.db.Get(parts[1])

	if errors.Is(err, database.ErrNotFound) {
		fmt.Fprintln(r.out, "NULL")
		return
	}

	if err != nil {
		r.printError(err.Error())
		return
	}

	fmt.Fprintln(r.out, value)
}

func (r *Runner) executeDelete(parts []string) {
	if len(parts) != 2 {
		r.printError("DEL requires 1 argument")
		return
	}

	deleted, err := r.db.Delete(parts[1])
	if err != nil {
		r.printError(err.Error())
		return
	}

	if deleted {
		fmt.Fprintln(r.out, "1")
	} else {
		fmt.Fprintln(r.out, "0")
	}
}

func (r *Runner) executeExists(parts []string) {
	if len(parts) != 2 {
		r.printError("EXISTS requires 1 argument")
		return
	}

	if r.db.Exists(parts[1]) {
		fmt.Fprintln(r.out, "1")
	} else {
		fmt.Fprintln(r.out, "0")
	}
}

func (r *Runner) executeMSet(parts []string) {
	args := parts[1:]

	if len(args) == 0 || len(args)%2 != 0 {
		r.printError("MSET requires key-value pairs")
		return
	}

	for i := 0; i < len(args); i += 2 {
		key := args[i]
		value := args[i+1]

		if err := r.db.Set(key, value); err != nil {
			r.printError(err.Error())
			return
		}
	}

	fmt.Fprintln(r.out, "OK")
}

func (r *Runner) executeMGet(parts []string) {
	if len(parts) < 2 {
		r.printError("MGET requires at least 1 key")
		return
	}

	for _, key := range parts[1:] {
		value, err := r.db.Get(key)

		if errors.Is(err, database.ErrNotFound) {
			fmt.Fprintln(r.out, "NULL")
			continue
		}

		if err != nil {
			r.printError(err.Error())
			return
		}

		fmt.Fprintln(r.out, value)
	}
}

func (r *Runner) executeFlushDB(parts []string) {
	if len(parts) != 1 {
		r.printError("FLUSHDB does not accept arguments")
		return
	}

	if err := r.db.FlushDB(); err != nil {
		r.printError(err.Error())
		return
	}

	fmt.Fprintln(r.out, "OK")
}

func (r *Runner) printError(message string) {
	fmt.Fprintf(r.out, "ERROR %s\n", message)
}
