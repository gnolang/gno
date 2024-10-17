package main

type SelfReferencing interface {
	SelfReferencing
}

func main() {
	println("ok")
}

// Error:
// main/files/recursive6a.gno:1:1: invalid recursive type: SelfReferencing -> SelfReferencing
