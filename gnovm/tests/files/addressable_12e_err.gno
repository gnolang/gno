package main

type T struct{ G int }

func (t *T) SetG(v int) { t.G = v }

type S struct {
	F int
	T T
}

func main() {
	var x any = [2]int{1, 1}
	x.([2]int)[0] = 9
	println(x.([2]int)[0])
}

// Error:
// main/addressable_12e_err.gno:14:2-19: cannot assign to x<VPBlock(1,0)>.([(const (2 int))](const-type int))[(const (0 int))]

// TypeCheckError:
// main/addressable_12e_err.gno:14:2: cannot assign to x.([2]int)[0] (neither addressable nor a map index expression)
