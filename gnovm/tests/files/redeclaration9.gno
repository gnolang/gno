package main

type a struct{}

func (_ a) method() int { return 1 }

func (_ a) method() int { return 2 }

func main() {
	println("hello")
}

// Error:
// running package "main": redeclarations for identifiers: "a<VPUverse(0)>.method"
