package main

import (
	"errors"
	"fmt"
)

// sentinel errors
var (
	ErrNotFound = errors.New("not found")
	ErrTimeout  = errors.New("timeout")
)

// wrappedError wraps another error
type wrappedError struct {
	msg string
	err error
}

func (e *wrappedError) Error() string { return e.msg + ": " + e.err.Error() }
func (e *wrappedError) Unwrap() error { return e.err }

func main() {
	// Test errors.Is with direct match
	err := ErrNotFound
	fmt.Println("Is(ErrNotFound, ErrNotFound):", errors.Is(err, ErrNotFound))
	fmt.Println("Is(ErrNotFound, ErrTimeout):", errors.Is(err, ErrTimeout))

	// Test errors.Is with wrapped error
	wrapped := &wrappedError{"operation failed", ErrNotFound}
	fmt.Println("Is(wrapped, ErrNotFound):", errors.Is(wrapped, ErrNotFound))
	fmt.Println("Is(wrapped, ErrTimeout):", errors.Is(wrapped, ErrTimeout))

	// Test errors.Is with nil
	fmt.Println("Is(nil, nil):", errors.Is(nil, nil))
	fmt.Println("Is(err, nil):", errors.Is(err, nil))

	// Test errors.Unwrap
	fmt.Println("Unwrap(wrapped):", errors.Unwrap(wrapped))
	fmt.Println("Unwrap(simple):", errors.Unwrap(ErrNotFound))

	// Test errors.Join
	joined := errors.Join(ErrNotFound, ErrTimeout)
	fmt.Println("Join Error:", joined.Error())
	fmt.Println("Is(joined, ErrNotFound):", errors.Is(joined, ErrNotFound))
	fmt.Println("Is(joined, ErrTimeout):", errors.Is(joined, ErrTimeout))

	// Test errors.Join with nil filtering
	joinedWithNil := errors.Join(nil, ErrNotFound, nil)
	fmt.Println("Join with nils:", joinedWithNil.Error())

	// Test errors.Join returns nil if all nil
	allNil := errors.Join(nil, nil)
	fmt.Println("Join all nil:", allNil)
}

// Output:
// Is(ErrNotFound, ErrNotFound): true
// Is(ErrNotFound, ErrTimeout): false
// Is(wrapped, ErrNotFound): true
// Is(wrapped, ErrTimeout): false
// Is(nil, nil): true
// Is(err, nil): false
// Unwrap(wrapped): not found
// Unwrap(simple): <nil>
// Join Error: not found
// timeout
// Is(joined, ErrNotFound): true
// Is(joined, ErrTimeout): true
// Join with nils: not found
// Join all nil: <nil>
