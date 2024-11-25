package main

type (
	nat  []word
	word int
)

func main() {
	var a nat
	b := []word{0}
	a = b

	println(a)
}

// Output:
// (slice[(0 main.word)] main.nat)
