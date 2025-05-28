package main

type s struct {
	x int
}

func main() {
	s := s{1}
	const v = s.x
}

// Error:
// main/const36.gno:9:8-15: s<VPBlock(1,0)>.x (variable of type int) is not constant

// TypeCheckError:
// main/const36.gno:9:12: s.x (variable of type int) is not constant
