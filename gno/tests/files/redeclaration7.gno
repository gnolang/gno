package main

func f1() int { return 1 }

func f1() int { return 2 }

func main() {
	println("hello", f1())
}

// Error:
// files/redeclaration7.gno:5:6: f1 redeclared in this block
// 	previous declaration at files/redeclaration7.gno:3:6
