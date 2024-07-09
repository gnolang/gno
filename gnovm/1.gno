package main

const a = b + 1
const b = a

func main() {
	println(a)
	println(b)
}

// Error:
// main/1.gno:3: 2 constant definition loop with a
