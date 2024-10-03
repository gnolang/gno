package main

import (
	"time"
)

func main() {
	var t time.Time
	println(t.String())
}

// Error:
// ../_test/redeclaration_global2.gno:5:2: time/redeclaration_global2.gno redeclared in this block
