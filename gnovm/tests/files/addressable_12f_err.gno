package main

type T struct{ G int }

func (t *T) SetG(v int) { t.G = v }

type S struct {
	F int
	T T
}

func main() {
	var x any = S{F: 1}
	x.(S).T.SetG(9)
	println(x.(S).T.G)
}

// Error:
// main/addressable_12f_err.gno:14:2-14: cannot call pointer method SetG on main.T

// TypeCheckError:
// main/addressable_12f_err.gno:14:10: cannot call pointer method SetG on T
