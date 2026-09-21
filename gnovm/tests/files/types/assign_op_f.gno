// Blank target wins over untyped-nil RHS in compound assign, like go/types.
package main

func main() {
	_ += nil
}

// Error:
// main/assign_op_f.gno:5:2-10: cannot use _ as value or type

// TypeCheckError:
// main/assign_op_f.gno:5:2: cannot use _ as value or type
