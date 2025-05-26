package main

func main() {
	type S struct {
		T S
	}

	var a, b S
	println(a == b)
}

// Error:
// main/recursive1f.gno:4:7-6:3: invalid recursive type: S -> S

// TypeCheckError:
// main/recursive1f.gno:4:7: invalid recursive type: S refers to itself
