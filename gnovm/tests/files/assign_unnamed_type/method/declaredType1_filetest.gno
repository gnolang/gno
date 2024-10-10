package main

type (
	word uint
	nat  []word
)

// structLit
type Int struct {
	neg bool
	abs nat
}

func (n nat) add() bool {
	return true
}

func main() {
	z := &Int{
		neg: true,
		abs: []word{0},
	}
	println(z.abs.add())
}

// Output:
// true
