package main

type T struct{ G int }

func (t *T) SetG(v int) { t.G = v }

type S struct {
	F int
	T T
}

func main() {
	var arr [1]any
	arr[0] = S{F: 1}
	arr2 := arr
	arr2[0].(S).F = 9
	println(arr[0].(S).F, arr2[0].(S).F)
}

// Error:
// main/addressable_12a_err.gno:16:2-19: cannot assign to arr2<VPBlock(1,1)>[(const (0 int))].(typeval{main.S}).F

// TypeCheckError:
// main/addressable_12a_err.gno:16:2: cannot assign to arr2[0].(S).F (neither addressable nor a map index expression)
