package main

type a struct{}

func (_ a) method() int { return 1 }

func (_ a) method() int { return 2 }

func main() {
	println("hello")
}

// Error:
// main/files/redeclaration9.gno:7: redeclaration of method a.method
