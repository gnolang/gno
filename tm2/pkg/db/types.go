package db

// A DB is goroutine safe.
//
// NOTE(morgan): these methods used to have comments to describe the semantics;
// but the semantics described were wrong, and have been removed to avoid doing
// more harm than good. Inspect the implementations and tests to understand
// actual details, and maybe update this documentation for the next person who
// has to deal with this interface.
type DB interface {
	// CONTRACT: key, value readonly []byte
	Get([]byte) []byte

	// CONTRACT: key, value readonly []byte
	Has(key []byte) bool

	// CONTRACT: key, value readonly []byte
	Set([]byte, []byte)
	SetSync([]byte, []byte)

	// CONTRACT: key readonly []byte
	Delete([]byte)
	DeleteSync([]byte)

	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end readonly []byte
	Iterator(start, end []byte) Iterator

	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end readonly []byte
	ReverseIterator(start, end []byte) Iterator

	// Closes the connection.
	Close()

	// Creates a batch for atomic updates.
	NewBatch() Batch

	// For debugging
	Print()

	// Stats returns a map of property values for all keys and the size of the cache.
	Stats() map[string]string
}

//----------------------------------------
// Batch

// Batch Close must be called when the program no longer needs the object.
type Batch interface {
	SetDeleter
	Write()
	WriteSync()
	Close()
}

type SetDeleter interface {
	Set(key, value []byte) // CONTRACT: key, value readonly []byte
	Delete(key []byte)     // CONTRACT: key readonly []byte
}

//----------------------------------------
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
type Iterator interface {
	// The start & end (exclusive) limits to iterate over.
	// If end < start, then the Iterator goes in reverse order.
	//
	// A domain of ([]byte{12, 13}, []byte{12, 14}) will iterate
	// over anything with the prefix []byte{12, 13}.
	//
	// The smallest key is the empty byte array []byte{} - see BeginningKey().
	// The largest key is the nil byte array []byte(nil) - see EndingKey().
	// CONTRACT: start, end readonly []byte
	Domain() (start []byte, end []byte)

	// Valid returns whether the current position is valid.
	// Once invalid, an Iterator is forever invalid.
	Valid() bool

	// Next moves the iterator to the next sequential key in the database, as
	// defined by order of iteration.
	//
	// If Valid returns false, this method will panic.
	Next()

	// Key returns the key of the cursor.
	// If Valid returns false, this method will panic.
	// CONTRACT: key readonly []byte
	Key() (key []byte)

	// Value returns the value of the cursor.
	// If Valid returns false, this method will panic.
	// CONTRACT: value readonly []byte
	Value() (value []byte)

	// Close releases the Iterator.
	Close()
}
