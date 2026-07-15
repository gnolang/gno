package main

// Regression test: a backward goto that jumps out of deeply nested for loops
// used to panic in Machine.GotoJump with "slice bounds out of range [:-1]".
// The stmt-stack truncation subtracted the crossed frame count twice, which
// underflowed once the number of crossed loop frames exceeded the number of
// enclosing bodyStmts (here, three nested loops jumping back to a label in the
// function body).

func main() {
	i := 0
top:
	println("i =", i)
	for a := 0; a < 2; a++ {
		for b := 0; b < 2; b++ {
			for c := 0; c < 2; c++ {
				i++
				if i < 3 {
					goto top
				}
			}
		}
	}
	println("done", i)
}

// Output:
// i = 0
// i = 1
// i = 2
// done 10
