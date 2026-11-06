package main

type T struct{ G int }

func (t *T) SetG(v int) { t.G = v }

type S struct {
	F int
	T T
}

func main() {
	var x any = S{F: 1}
	f := x.(S).T.SetG
	f(9)
	println(x.(S).T.G)
}

// Error:
// main/addressable_12g_err.gno:14:7-19: cannot call pointer method SetG on main.T

// TypeCheckError:
// main/addressable_12g_err.gno:14:15: cannot call pointer method SetG on T
