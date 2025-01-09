package main

type s struct {
	x int
}

func main() {
	s := s{1}
	const v = s.x
}

// Error:
// main/files/const36.gno:9:8: unexpected selector expression type *gnolang.DeclaredType
