package main

func main() {
	x := 5
	x
}

// Error:
// main/expr_stmt1.gno:5:2-3: invalid operation: x<VPBlock(1,0)> is not used

// TypeCheckError:
// main/expr_stmt1.gno:5:2: x (variable of type int) is not used
