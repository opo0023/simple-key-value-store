package persistence

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
)

// Record represents one operation in the append-only log.
type Record struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// Log manages the append-only data file.
type Log struct {
	file *os.File
	path string
}

// Open opens or creates the database log file.
func Open(path string) (*Log, error) {
	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0o600,
	)
	if err != nil {
		return nil, err
	}

	return &Log{
		file: file,
		path: path,
	}, nil
}

// Append writes one record to disk immediately.
func (l *Log) Append(record Record) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	if _, err := l.file.Write(data); err != nil {
		return err
	}

	// Sync forces the append to stable storage before the
	// operation is considered successfully persisted.
	return l.file.Sync()
}

// Replay reads each persisted record from the beginning of the log.
func (l *Log) Replay(apply func(Record) error) error {
	file, err := os.Open(l.path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		line, readErr := reader.ReadBytes('\n')

		if len(line) > 0 {
			var record Record

			if err := json.Unmarshal(line, &record); err != nil {
				return err
			}

			if err := apply(record); err != nil {
				return err
			}
		}

		if errors.Is(readErr, io.EOF) {
			break
		}

		if readErr != nil {
			return readErr
		}
	}

	return nil
}

// Truncate removes all records and reopens the log file.
//
// The active file handle is closed first because Windows may reject
// truncation while the same process still holds the file open.
func (l *Log) Truncate() error {
	if l.file != nil {
		if err := l.file.Close(); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(
		l.path,
		os.O_CREATE|os.O_RDWR|os.O_TRUNC|os.O_APPEND,
		0o600,
	)
	if err != nil {
		return err
	}

	l.file = file

	return l.file.Sync()
}

// Close closes the underlying log file.
func (l *Log) Close() error {
	if l.file == nil {
		return nil
	}

	return l.file.Close()
}
