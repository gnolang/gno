package db

// DBs are goroutine safe.
type DB interface {
	// Get returns nil iff key doesn't exist.
	// A nil key is interpreted as an empty byteslice.
	// CONTRACT: key, value readonly []byte
	Get([]byte) ([]byte, error)

	// Has checks if a key exists.
	// A nil key is interpreted as an empty byteslice.
	// CONTRACT: key, value readonly []byte
	Has(key []byte) (bool, error)

	// Set sets the key.
	// A nil key is interpreted as an empty byteslice.
	// CONTRACT: key, value readonly []byte
	Set([]byte, []byte) error
	SetSync([]byte, []byte) error

	// Delete deletes the key.
	// A nil key is interpreted as an empty byteslice.
	// CONTRACT: key readonly []byte
	Delete([]byte) error
	DeleteSync([]byte) error

	// Iterate over a domain of keys in ascending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// A nil start is interpreted as an empty byteslice.
	// If end is nil, iterates up to the last item (inclusive).
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end readonly []byte
	Iterator(start, end []byte) (Iterator, error)

	// Iterate over a domain of keys in descending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// If start is nil, iterates up to the first/least item (inclusive).
	// If end is nil, iterates from the last/greatest item (inclusive).
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end readonly []byte
	ReverseIterator(start, end []byte) (Iterator, error)

	// Closes the connection.
	Close() error

	// Creates a batch for atomic updates.
	NewBatch() Batch

	// NewBatchWithSize create a new batch for atomic updates, but with pre-allocated size.
	// This will does the same thing as NewBatch if the batch implementation doesn't support pre-allocation.
	NewBatchWithSize(int) Batch

	// For debugging
	Print() error

	// Stats returns a map of property values for all keys and the size of the cache.
	Stats() map[string]string
}

// ----------------------------------------
// Batch

// Batch represents a group of writes. They may or may not be written atomically depending on the
// backend. Callers must call Close on the batch when done.
//
// Batch Close must be called when the program no longer needs the object.
type Batch interface {
	// Set sets a key/value pair.
	// CONTRACT: key, value readonly []byte
	Set(key, value []byte) error

	// Delete deletes a key/value pair.
	// CONTRACT: key readonly []byte
	Delete(key []byte) error

	// Write writes the batch, possibly without flushing to disk. Only Close() can be called after,
	// other methods will error.
	Write() error

	// WriteSync writes the batch and flushes it to disk. Only Close() can be called after, other
	// methods will error.
	WriteSync() error

	// Close closes the batch. It is idempotent, but calls to other methods afterwards will error.
	Close() error

	// GetByteSize that returns the current size of the batch in bytes. Depending on the implementation,
	// this may return the size of the underlying LSM batch, including the size of additional metadata
	// on top of the expected key and value total byte count.
	GetByteSize() (int, error)
}

// ----------------------------------------
// Iterator

/*
Usage:

var itr Iterator = ...
defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		k, v := itr.Key(); itr.Value()
		// ...
	}
*/
// Iterator represents an iterator over a domain of keys. Callers must call
// Close when done. No writes can happen to a domain while there exists an
// iterator over it. Some backends may take out database locks to ensure this
// will not happen.
//
// Callers must make sure the iterator is valid before calling any methods on it,
// otherwise these methods will panic.
type Iterator interface {
	// Domain returns the start (inclusive) and end (exclusive) limits of the iterator.
	Domain() (start, end []byte)

	// Valid returns whether the current iterator is valid. Once invalid, the Iterator remains
	// invalid forever.
	Valid() bool

	// Next moves the iterator to the next key in the database, as defined by order of iteration.
	// If Valid returns false, this method will panic.
	Next()

	// Key returns the key at the current position. Panics if the iterator is invalid.
	// Note, the key returned should be a copy and thus safe for modification.
	Key() []byte

	// Value returns the value at the current position. Panics if the iterator is
	// invalid.
	// Note, the value returned should be a copy and thus safe for modification.
	Value() []byte

	// Error returns the last error encountered by the iterator, if any.
	Error() error

	// Close closes the iterator, releasing any allocated resources.
	Close() error
}
