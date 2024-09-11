package main

type S struct {
	A [2][]S
}

func main() {
	var a, b S
	println(a)
}

// Output:
// (struct{(array[(nil []main.S),(nil []main.S)] [2][]main.S)} main.S)
