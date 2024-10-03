package main

func f() func() {
	return nil
}

func main() {
	g := f()
	println(g)
	if g == nil {
		println("nil func")
	}
}

// Output:
// nil func()()
// nil func
