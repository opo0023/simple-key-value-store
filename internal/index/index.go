package index

import "time"

// Entry represents one key-value pair in the database.
type Entry struct {
	Key       string
	Value     string
	ExpiresAt int64
	Next      *Entry
}

// Index stores entries using a custom linked list.
type Index struct {
	Head *Entry
}

// New creates an empty index.
func New() *Index {
	return &Index{}
}

// Set creates or replaces a key-value pair.
//
// Replacing a key removes its previous expiration.
func (i *Index) Set(key string, value string) {
	entry, exists := i.findEntry(key)

	if exists {
		entry.Value = value
		entry.ExpiresAt = 0
		return
	}

	i.Head = &Entry{
		Key:   key,
		Value: value,
		Next:  i.Head,
	}
}

// Get returns the value for a nonexpired key.
func (i *Index) Get(key string) (string, bool) {
	var previous *Entry
	current := i.Head
	now := time.Now().Unix()

	for current != nil {
		if current.Key == key {
			if isExpired(current, now) {
				i.removeEntry(previous, current)
				return "", false
			}

			return current.Value, true
		}

		previous = current
		current = current.Next
	}

	return "", false
}

// Delete removes a key.
func (i *Index) Delete(key string) bool {
	var previous *Entry
	current := i.Head

	for current != nil {
		if current.Key == key {
			i.removeEntry(previous, current)
			return true
		}

		previous = current
		current = current.Next
	}

	return false
}

// Exists reports whether a nonexpired key exists.
func (i *Index) Exists(key string) bool {
	_, exists := i.Get(key)
	return exists
}

// SetExpiration assigns an absolute Unix expiration timestamp to a key.
func (i *Index) SetExpiration(key string, expiresAt int64) bool {
	entry, exists := i.findEntry(key)
	if !exists {
		return false
	}

	if isExpired(entry, time.Now().Unix()) {
		i.Delete(key)
		return false
	}

	entry.ExpiresAt = expiresAt
	return true
}

// TTL returns the remaining lifetime of a key.
//
// -2 means the key does not exist.
// -1 means the key exists but has no expiration.
func (i *Index) TTL(key string) int64 {
	entry, exists := i.findEntry(key)
	if !exists {
		return -2
	}

	if entry.ExpiresAt == 0 {
		return -1
	}

	remaining := entry.ExpiresAt - time.Now().Unix()

	if remaining <= 0 {
		i.Delete(key)
		return -2
	}

	return remaining
}

// Clear removes all entries.
func (i *Index) Clear() {
	i.Head = nil
}

func (i *Index) findEntry(key string) (*Entry, bool) {
	current := i.Head

	for current != nil {
		if current.Key == key {
			return current, true
		}

		current = current.Next
	}

	return nil, false
}

func (i *Index) removeEntry(previous *Entry, current *Entry) {
	if previous == nil {
		i.Head = current.Next
		return
	}

	previous.Next = current.Next
}

func isExpired(entry *Entry, now int64) bool {
	return entry.ExpiresAt > 0 && entry.ExpiresAt <= now
}
