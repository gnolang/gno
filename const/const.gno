package main

type f func()

func foo() {
	println("foo")
}

func bar() int {
	println("bar")
	return 0
}

func main() {
	const x = len("qSDqs")
	println(x)
}