package main

func main() {
	type S struct {
		T S
	}

	var a, b S
	println(a == b)
}

// Error:
// main/files/recursive1f.gno:4:7: invalid recursive type: S -> S

// TypeCheckError:
// main/files/recursive1f.gno:4:7: invalid recursive type: S refers to itself
