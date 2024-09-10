package main

type Int int8

func (i Int) Inc() {
	i++
}

// right is typed const, can not use as its correspondence declared type
func main() {
	var a Int
	a = int8(0)
	a.Inc()
	println(a)
}

// Error:
// main/files/types/incdec_a4.gno:12:2: cannot use int8 as main.Int without explicit conversion
