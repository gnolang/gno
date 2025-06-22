package main

type MyStruct struct {
	interface {
			B()
		}
}

func main() {
	println("hello")
}

// Error:
// struct62.gno:4:2: expected '}', found 'interface'

// TypeCheckError:
// main/struct62.gno:4:2: expected '}', found 'interface'
