package main

func anotherRecover() {
	if r := recover(); r != nil {
		println(r)
	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			println(r)
		}
	}()
	defer anotherRecover()
	defer func() {
		if r := recover(); r != nil {
			panic("panic in defer func")
		}
	}()

	panic("panic in main")
}

// Output:
// panic in defer func
