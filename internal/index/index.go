package index

// Entry represents one key-value pair in the database.
type Entry struct {
	Key   string
	Value string
	Next  *Entry
}

// Index stores entries using a custom linked list.
//
// A built-in Go map is intentionally not used because the project
// requires a custom in-memory index structure.
type Index struct {
	Head *Entry
}

// New creates an empty index.
func New() *Index {
	return &Index{}
}

// Set stores a key-value pair.
//
// If the key already exists, its value is replaced.
// This provides "last write wins" behavior.
func (i *Index) Set(key string, value string) {
	current := i.Head

	for current != nil {
		if current.Key == key {
			current.Value = value
			return
		}

		current = current.Next
	}

	newEntry := &Entry{
		Key:   key,
		Value: value,
		Next:  i.Head,
	}

	i.Head = newEntry
}

// Get returns the value associated with a key.
func (i *Index) Get(key string) (string, bool) {
	current := i.Head

	for current != nil {
		if current.Key == key {
			return current.Value, true
		}

		current = current.Next
	}

	return "", false
}

// Delete removes a key from the index.
func (i *Index) Delete(key string) bool {
	var previous *Entry
	current := i.Head

	for current != nil {
		if current.Key == key {
			if previous == nil {
				i.Head = current.Next
			} else {
				previous.Next = current.Next
			}

			return true
		}

		previous = current
		current = current.Next
	}

	return false
}

// Exists reports whether a key exists.
func (i *Index) Exists(key string) bool {
	_, exists := i.Get(key)
	return exists
}

// Clear removes all entries.
func (i *Index) Clear() {
	i.Head = nil
}
