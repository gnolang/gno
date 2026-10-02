package main

func main() {
	x := 5
	x + 1
}

// Error:
// main/expr_stmt2.gno:5:2-7: invalid operation: x<VPBlock(1,0)> + (const (1 int)) is not used

// TypeCheckError:
// main/expr_stmt2.gno:5:2: x + 1 (value of type int) is not used
