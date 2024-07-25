package main

func main() {
	type A struct {
		X B
	}

	type B struct {
		X C
	}

	type C struct {
		X A
	}

	var p, q A
	println(p == q)
}

// Error:
// main/files/recursive2c.gno:3: name B not defined in fileset with files [files/recursive2c.gno]
