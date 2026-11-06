package main

type T struct{ G int }

func (t *T) SetG(v int) { t.G = v }

type S struct {
	F int
	T T
}

func main() {
	var x any = S{F: 1}
	x.(S).F += 8
	println(x.(S).F)
}

// Error:
// main/addressable_12b_err.gno:14:2-14: cannot assign to x<VPBlock(1,0)>.(typeval{main.S}).F

// TypeCheckError:
// main/addressable_12b_err.gno:14:2: cannot assign to x.(S).F (neither addressable nor a map index expression)
