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
	defer func() {
		// g() shouldn't be able to recover from f()'s panic, because the recover
		// is declared in a deferred closure that is absent from the stack at the
		// time of the panic.
		// See go behavior here https://go.dev/play/p/CcMGgY606O-
		// See Rob's examples https://groups.google.com/g/golang-nuts/c/HOXNBQu5c-Q/m/Je0qo1hbxIsJ
		println("g recover", recover())
	}()
}

// Output:
// g recover undefined
// f recover wtf
