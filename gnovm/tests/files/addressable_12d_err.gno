package main

type T struct{ G int }

func (t *T) SetG(v int) { t.G = v }

type S struct {
	F int
	T T
}

func main() {
	var x any = S{F: 1}
	for x.(S).F = range [3]int{} {
	}
	println(x.(S).F)
}

// Error:
// main/addressable_12d_err.gno:14:2-15:3: cannot assign to x<VPBlock(2,0)>.(typeval{main.S}).F

// TypeCheckError:
// main/addressable_12d_err.gno:14:6: cannot assign to x.(S).F (neither addressable nor a map index expression)
