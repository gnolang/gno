package main

import (
	"gno.land/r/demo/recursive"
)

func main() {
	recursive.Do()
}

// Error:
// main/gno.land/r/demo/recursive/z_0_filetest.gno:1: main.gno: Cyclic dependency detected: A (Line: 3) -> B (Line: 3) -> C (Line: 7) -> A (Line: 3)
