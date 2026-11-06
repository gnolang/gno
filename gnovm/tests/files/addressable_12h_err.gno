package main

type T struct{ G int }

func (t *T) SetG(v int) { t.G = v }

type S struct {
	F int
	T T
}

var arr = [1]any{S{F: 1}}

func get() S { return arr[0].(S) }

func main() {
	get().F = 9
	println(arr[0].(S).F)
}

// Error:
// main/addressable_12h_err.gno:17:2-13: cannot assign to get<VPBlock(3,3)>().F

// TypeCheckError:
// main/addressable_12h_err.gno:17:2: cannot assign to get().F (neither addressable nor a map index expression)
