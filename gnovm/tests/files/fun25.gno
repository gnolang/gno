package main

func f(x string) (a int, b int) { return x, 5 }

func main() {
	print("hello")
}

// Error:
// main/fun25.gno:3:35-46: cannot use string as int

// TypeCheckError:
// main/fun25.gno:3:42: cannot use x (variable of type string) as int value in return statement
