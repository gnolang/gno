package db

import (
	corestore "cosmossdk.io/core/store"
)

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

// Batch Close must be called when the program no longer needs the object.
type Batch = corestore.Batch

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
type Iterator = corestore.Iterator
