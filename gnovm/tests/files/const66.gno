package main

const c = 1

func main() {
	{
		c = 2
		c := 3
		_ = c
	}
}

// Error:
// main/const66.gno:7:3-8: cannot assign to const c

// TypeCheckError:
// main/const66.gno:7:3: cannot assign to c (neither addressable nor a map index expression)
