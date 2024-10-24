package main

type s struct {
	x int
}

func main() {
	s := s{1}
	const v = s.x
}

// Error:
// main/files/const36.gno:9:8: s<VPBlock(1,0)>.x (variable of type main.s) is not constant
