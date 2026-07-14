package database

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/opo0023/simple-key-value-store/internal/index"
	"github.com/opo0023/simple-key-value-store/internal/persistence"
)

var (
	ErrNotFound   = errors.New("key not found")
	ErrNotInteger = errors.New("value is not an integer")
)

// RangeEntry represents one result from a range query.
type RangeEntry struct {
	Key   string
	Value string
}

// Database connects the in-memory index to the append-only log.
type Database struct {
	index *index.Index
	log   *persistence.Log
}

// Open opens the database and replays persisted operations.
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

// Set stores and persists a key-value pair.
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

// Delete removes and persists a key.
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

// Increment changes a numeric key by the provided amount.
//
// Missing keys begin at zero.
func (db *Database) Increment(key string, amount int64) (int64, error) {
	current := int64(0)

	value, exists := db.index.Get(key)
	if exists {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, ErrNotInteger
		}

		current = parsed
	}

	result := current + amount

	if err := db.Set(key, strconv.FormatInt(result, 10)); err != nil {
		return 0, err
	}

	return result, nil
}

// Expire assigns a time-to-live to an existing key.
func (db *Database) Expire(key string, seconds int64) (bool, error) {
	if seconds < 0 {
		return false, errors.New("expiration must be nonnegative")
	}

	if !db.index.Exists(key) {
		return false, nil
	}

	expiresAt := time.Now().Unix() + seconds

	record := persistence.Record{
		Command: "EXPIRE_AT",
		Args: []string{
			key,
			strconv.FormatInt(expiresAt, 10),
		},
	}

	if err := db.log.Append(record); err != nil {
		return false, err
	}

	db.index.SetExpiration(key, expiresAt)
	return true, nil
}

// TTL returns the remaining lifetime of a key.
//
// -2 means the key does not exist.
// -1 means the key exists without an expiration.
func (db *Database) TTL(key string) int64 {
	return db.index.TTL(key)
}

// Range returns keys between start and end, inclusive.
//
// Results are sorted lexicographically by key.
func (db *Database) Range(start string, end string) []RangeEntry {
	if start > end {
		return []RangeEntry{}
	}

	entries := db.index.Entries()
	results := make([]RangeEntry, 0)

	for _, entry := range entries {
		if entry.Key < start || entry.Key > end {
			continue
		}

		results = append(results, RangeEntry{
			Key:   entry.Key,
			Value: entry.Value,
		})
	}

	sort.Slice(results, func(left int, right int) bool {
		return results[left].Key < results[right].Key
	})

	return results
}

// FlushDB removes all keys and empties the persistence log.
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

	case "EXPIRE_AT":
		if len(record.Args) != 2 {
			return fmt.Errorf("invalid EXPIRE_AT record")
		}

		expiresAt, err := strconv.ParseInt(record.Args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid expiration timestamp: %w", err)
		}

		db.index.SetExpiration(record.Args[0], expiresAt)

	default:
		return fmt.Errorf("unknown persisted command: %s", record.Command)
	}

	return nil
}
