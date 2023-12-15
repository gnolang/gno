package main

// operation is an enumeration type representing different types of operations. Used in diff algorithm
// to indicates differences between files.
type operation uint

const (
	// insert represents an insertion operation.
	insert operation = iota + 1
	// delete represents a deletion operation.
	delete
	// equal represents an equal operation.
	equal
)

// String returns a string representation of the operation.
func (op operation) String() string {
	switch op {
	case insert:
		return "INS"
	case delete:
		return "DEL"
	case equal:
		return "EQ"
	default:
		return "UNKNOWN"
	}
}
