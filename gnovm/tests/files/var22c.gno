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
// main/files/var22c.gno:10:6: missing init expr for z<!VPUverse(0)>