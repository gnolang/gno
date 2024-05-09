package main

import (
	"gno.land/r/demo/recursive"
)

func main() {
	recursive.Do()
}

// Error:
// main/gno.land/r/demo/recursive/z_0_filetest.gno:1: Cyclic dependency detected: A(File: main.gno, Line: 3) -> B(File: foo.gno, Line: 3) -> C(File: foo.gno, Line: 7) -> A(File: main.gno, Line: 3)
