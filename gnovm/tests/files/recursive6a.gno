package main

type SelfReferencing interface {
	SelfReferencing
}

func main() {
	println("ok")
}

// Error:
// main/files/recursive6a.gno:3:6: invalid recursive type: SelfReferencing -> SelfReferencing

// TypeCheckError:
// main/files/recursive6a.gno:3:6: invalid recursive type: SelfReferencing refers to itself
