package main

type T struct{ n int }

func (t *T) Inc() { t.n++ }

func main() {
	// A method expression is a func value. (*T) here is a pointer type
	// expression, not an indirection of anything.
	_ = &(*T).Inc
}

// TypeCheckError:
// main/addressable_12e_err.gno:10:7: invalid operation: cannot take address of (*T).Inc (value of type func(*T))

// Error:
// main/addressable_12e_err.gno:10:6-15: invalid operation: cannot take address of *(typeval{main.T}).Inc (value of type func(*main.T))
