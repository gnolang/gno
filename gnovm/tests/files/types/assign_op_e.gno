// A no-value call as the compound-assign RHS.
package main

func f() {}

func main() {
	i := 1
	i += f()
	_ = i
}

// Error:
// main/assign_op_e.gno:8:2-10: f<VPBlock(3,0)> (no value) used as value

// TypeCheckError:
// main/assign_op_e.gno:8:7: f() (no value) used as value
