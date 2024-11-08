package main

func main() {
	a, b := 1, 2
	a, b := 3, 4
}

// Error:
// files/redefine6.gno:5:2: no new variables on left side of :=
