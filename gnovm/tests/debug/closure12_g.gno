package main

func main() {
	f := func(a int) bool {
		println(a)
		return true
	}

	println(f(5))
}

// Output:
// 5
// true
