// Order pin: invalid RHS (no-value call) wins over unassignable target
// (const), like go/types.
package main

const c int = 5

func f() {}

func main() {
	c += f()
}

// Error:
// main/assign_op_g.gno:10:2-10: f<VPBlock(3,1)> (no value) used as value

// TypeCheckError:
// main/assign_op_g.gno:10:7: f() (no value) used as value
