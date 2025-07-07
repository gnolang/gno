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
// main/recursive2c.gno:4:7-6:3: name B not defined in fileset with files [recursive2c.gno]

// TypeCheckError:
// main/recursive2c.gno:5:5: undefined: B; main/recursive2c.gno:9:5: undefined: C
