package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/opo0023/simple-key-value-store/internal/cli"
	"github.com/opo0023/simple-key-value-store/internal/database"
)

func main() {
	db, err := database.Open("data.db")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to open database:", err)
		os.Exit(1)
	}

	defer func() {
		if err := db.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "failed to close database:", err)
		}
	}()

	runner := cli.NewRunner(db, os.Stdout)
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		if runner.Execute(scanner.Text()) {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to read input:", err)
		os.Exit(1)
	}
}
