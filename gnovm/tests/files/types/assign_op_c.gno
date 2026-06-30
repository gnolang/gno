// Blank as compound-assign target: `_ += x` desugars to `_ = _ + x`, using
// `_` as a value — illegal, unlike plain `_ = x`.
package main

func main() {
	_ += 1
	println("ok")
}

// Error:
// main/assign_op_c.gno:6:2-8: cannot use _ as value or type

// TypeCheckError:
// main/assign_op_c.gno:6:2: cannot use _ as value or type
