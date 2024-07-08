package main

func main() {
	f()
}

func f() {
	defer func() { println("f recover", recover()) }()
	defer g()
	panic("wtf")
}

func g() {
	println("g recover", recover())
}

// Output:
// g recover wtf
// f recover undefined
