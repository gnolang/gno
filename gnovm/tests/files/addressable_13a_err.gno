package main

type T struct{ n int }

func (t *T) Inc() { t.n++ }

func main() {
	m := map[string]T{"k": {1}}
	m["k"].Inc()
}

// TypeCheckError:
// main/addressable_13a_err.gno:9:9: cannot call pointer method Inc on T

// Error:
// main/addressable_13a_err.gno:9:2-12: cannot call pointer method Inc on main.T
