package main

func Bar() string {
	return
}

func main() {
	println(Bar())
}

// Error:
// 4:2: expected 1 return values; got 0
