package main

func main() {
	a := 1
	r := a << 1 // NOTE: go vet would fail, but still process
	println(r)
}

// Output:
// 2
