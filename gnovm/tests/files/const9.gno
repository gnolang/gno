package main

const (
	a = 2
	b = c + d
	c = a + d
	d = e + f
	e = b + 2
	f = 4
)

func main() {
	println(b)
}

// Error:
// main/const9.gno:5:2-11: invalid recursive value: b -> c -> d -> e -> b

// TypeCheckError:
// main/const9.gno:6:2: initialization cycle for c; main/const9.gno:6:2: 	c refers to d; main/const9.gno:7:2: 	d refers to e; main/const9.gno:8:2: 	e refers to b; main/const9.gno:5:2: 	b refers to c; main/const9.gno:5:2: initialization cycle for b; main/const9.gno:5:2: 	b refers to c; main/const9.gno:6:2: 	c refers to d; main/const9.gno:7:2: 	d refers to e; main/const9.gno:8:2: 	e refers to b
