package main

func main() {
	var t *int = nil
	println("pass")
}

// TypeCheckError:
// main/var33.gno:4:6: declared and not used: t

// Error:
// main/var33.gno:4:6-18': declared and not used: t
