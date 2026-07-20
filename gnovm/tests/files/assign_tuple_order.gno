package main

// Go spec: in tuple assignment, operands and RHS are evaluated first,
// then assignments happen left-to-right. A panic during the assignment
// of L_i must leave assignments to L_0..L_{i-1} in place.
// See https://github.com/golang/go/issues/23017.

func main() {
	m := map[int]int{}
	var p *int
	defer func() {
		recover()
		println(len(m), m[2])
	}()
	m[2], *p = 42, 2
}

// Output:
// 1 42
