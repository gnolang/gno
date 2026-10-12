// A type as the compound-assign RHS.
package main

func main() {
	i := 1
	i += int
	_ = i
}

// Error:
// main/assign_op_d.gno:6:2-10: int (type) is not an expression

// TypeCheckError:
// main/assign_op_d.gno:6:7: int (type) is not an expression
