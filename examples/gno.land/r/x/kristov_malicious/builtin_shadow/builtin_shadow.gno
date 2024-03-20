package main

func main() {
	// If shadowing disallowed, should result in panic concerning shadowing and not "foo" or "bar"
	panic("foo") 
}

func panic(s string) {
	print("bar")
}
