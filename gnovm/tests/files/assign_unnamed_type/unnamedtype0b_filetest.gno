package main

type (
	nat  []int
	word int
)

func main() {
	var a nat
	b := []word{0}
	a = b

	println(a)
}

// Error:
// main/unnamedtype0b.gno:11:2-7: cannot use []main.word as []int

// TypeCheckError:
// main/unnamedtype0b.gno:11:6: cannot use b (variable of type []word) as nat value in assignment
