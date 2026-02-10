package main

import "fmt"

func f() (a, b int) {
	return 1, 2
}

func main() {
	var x, y, z = 1, f()
	fmt.Println(x, y, z)
}

// Error:
// main/var22c.gno:10:6-22: missing init expr for z<!VPInvalid(0)>

// TypeCheckError:
// main/var22c.gno:10:12: missing init expr for z; main/var22c.gno:10:19: multiple-value f() (value of type (a int, b int)) in single-value context
