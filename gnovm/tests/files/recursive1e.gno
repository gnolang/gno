package main

type S struct {
	A [2][]S
}

func main() {
	var a, b S
	println(a)
}

// TypeCheckError:
// main/recursive1e.gno:8:9: declared and not used: b

// Error:
// main/recursive1e.gno:8:6-12': declared and not used: b
