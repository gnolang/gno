package main

type T struct{ n int }

func (t *T) Inc() { t.n++ }

func main() {
	m := map[string]T{"k": {1}}
	f := m["k"].Inc
	_ = f
}

// TypeCheckError:
// main/addressable_13c_err.gno:9:14: cannot call pointer method Inc on T

// Error:
// main/addressable_13c_err.gno:9:7-17: cannot call pointer method Inc on main.T
