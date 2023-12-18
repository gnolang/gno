package main

type word uint
type nat []word

// receiver
func (n nat) add() bool {
	return true
}

func Gen() interface{} {
	n := []word{0}
	return n
}

func Gen2() interface{} {
	n := nat([]word{0})
	return n
}

func main() {
	r := Gen()
	switch t := r.(type) {
	case nat:
		println("nat")
		println(t.add())
	case []word:
		println("[]word")
	default:
		println("should not happen")
	}

	r2 := Gen2()
	switch t := r2.(type) {
	case nat:
		println("nat")
		println(t.add())
	case []word:
		println("[]word")
	default:
		println("should not happen")
	}
}

// Output:
// []word
// nat
// true
