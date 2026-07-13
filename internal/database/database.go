package database

import (
	"errors"
	"fmt"

	"github.com/opo0023/simple-key-value-store/internal/index"
	"github.com/opo0023/simple-key-value-store/internal/persistence"
)

// ErrNotFound is returned when a key does not exist.
var ErrNotFound = errors.New("key not found")

// Database connects the custom in-memory index to the append-only log.
type Database struct {
	index *index.Index
	log   *persistence.Log
}

// Open opens the database and rebuilds the in-memory index from data.db.
func Open(path string) (*Database, error) {
	log, err := persistence.Open(path)
	if err != nil {
		return nil, err
	}

	db := &Database{
		index: index.New(),
		log:   log,
	}

	if err := db.log.Replay(db.applyRecord); err != nil {
		_ = db.log.Close()
		return nil, err
	}

	return db, nil
}

// Set persists and stores a key-value pair.
func (db *Database) Set(key string, value string) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	record := persistence.Record{
		Command: "SET",
		Args:    []string{key, value},
	}

	if err := db.log.Append(record); err != nil {
		return err
	}

	db.index.Set(key, value)
	return nil
}

// Get retrieves the value for a key.
func (db *Database) Get(key string) (string, error) {
	value, exists := db.index.Get(key)
	if !exists {
		return "", ErrNotFound
	}

	return value, nil
}

// Delete removes a key and persists the deletion.
func (db *Database) Delete(key string) (bool, error) {
	if !db.index.Exists(key) {
		return false, nil
	}

	record := persistence.Record{
		Command: "DEL",
		Args:    []string{key},
	}

	if err := db.log.Append(record); err != nil {
		return false, err
	}

	db.index.Delete(key)
	return true, nil
}

// Exists reports whether a key exists.
func (db *Database) Exists(key string) bool {
	return db.index.Exists(key)
}

// FlushDB clears the index and truncates the persistence log.
func (db *Database) FlushDB() error {
	if err := db.log.Truncate(); err != nil {
		return err
	}

	db.index.Clear()
	return nil
}

// Close closes the append-only log.
func (db *Database) Close() error {
	return db.log.Close()
}

// applyRecord applies one persisted operation during startup recovery.
func (db *Database) applyRecord(record persistence.Record) error {
	switch record.Command {
	case "SET":
		if len(record.Args) != 2 {
			return fmt.Errorf("invalid SET record")
		}

		db.index.Set(record.Args[0], record.Args[1])

	case "DEL":
		if len(record.Args) != 1 {
			return fmt.Errorf("invalid DEL record")
		}

		db.index.Delete(record.Args[0])

	default:
		return fmt.Errorf("unknown persisted command: %s", record.Command)
	}

	return nil
}
