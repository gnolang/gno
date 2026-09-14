package main

type T struct{ n int }

func (t *T) Inc() { t.n++ }

func mk() T { return T{1} }

func main() {
	mk().Inc()
}

// TypeCheckError:
// main/addressable_13b_err.gno:10:7: cannot call pointer method Inc on T

// Error:
// main/addressable_13b_err.gno:10:2-10: cannot call pointer method Inc on main.T
