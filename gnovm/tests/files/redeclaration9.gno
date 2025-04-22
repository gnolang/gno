package main

type a struct{}

func (_ a) method() int { return 1 }

func (_ a) method() int { return 2 }

func main() {
	println("hello")
}

// Error:
// main/files/redeclaration9.gno:7:1: redeclaration of method a.method

// TypeCheckError:
// main/files/redeclaration9.gno:7:12: method a.method already declared at main/files/redeclaration9.gno:5:12
