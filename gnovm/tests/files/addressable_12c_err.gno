package main

import "strconv"

func main() {
	_ = &strconv.Atoi
}

// TypeCheckError:
// main/addressable_12c_err.gno:6:7: invalid operation: cannot take address of strconv.Atoi (value of type func(s string) (int, error))

// Error:
// main/addressable_12c_err.gno:6:6-19: invalid operation: cannot take address of strconv<VPBlock(2,0)>.Atoi (value of type func(string) (int, .uverse.error))
