package vm

type (
	Word uint
	nat  []Word
)

var zero = &Int{
	neg: true,
	abs: []Word{0},
}

// structLit
type Int struct {
	neg bool
	abs nat
}

func GetZeroType() nat {
	a := zero.abs
	return a
}
