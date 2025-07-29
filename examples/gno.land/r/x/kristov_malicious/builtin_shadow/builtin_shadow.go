package main

func main() {
	panic("foo")
}

func panic(s string) {
	println("bar")
}
