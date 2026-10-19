package main

func main() {
	const t int
}

// TypeCheckError:
// main/const42.gno:4:8: missing init expr for t

// Error:
// main/const42.gno:4:8-13: missing init expr for t
