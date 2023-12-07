package main

// operation is an enumeration type representing different types of operations. Used in diff algorithm
// to indicates differences between files.
type operation uint

const (
	// INSERT represents an insertion operation.
	INSERT operation = 1
	// DELETE represents a deletion operation.
	DELETE operation = 2
	// MOVE represents a move operation.
	MOVE operation = 3
)

// String returns a string representation of the operation.
func (op operation) String() string {
	switch op {
	case INSERT:
		return "INS"
	case DELETE:
		return "DEL"
	case MOVE:
		return "MOV"
	default:
		return "UNKNOWN"
	}
}
