package main

type SelfReferencing interface {
	SelfReferencing
}

func main() {
	println("ok")
}

// Error:
// main/recursive6a.gno:3:6-5:2: invalid recursive type: SelfReferencing -> SelfReferencing

// TypeCheckError:
// main/recursive6a.gno:3:6: invalid recursive type: SelfReferencing refers to itself
