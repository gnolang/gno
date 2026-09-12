package main

type T struct{ n int }

func (t *T) Inc() { t.n++ }

type O struct{ in T }

func main() {
	O{}.in.Inc()
}

// TypeCheckError:
// main/addressable_13d_err.gno:10:9: cannot call pointer method Inc on T

// Error:
// main/addressable_13d_err.gno:10:2-12: cannot call pointer method Inc on main.T
