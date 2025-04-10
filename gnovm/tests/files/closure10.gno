package main

func trymake() func() {
	if s := "hello"; true {
		return func() {
			println(s)
		}
	}
	return func() {}
}

func main() {
	trymake()()
}

// Output:
// hello
