package main

// The receiver is reported as the type written at the call site, not the
// embedded type the method was promoted from.

type C struct{ n int }

func (c *C) Inc() { c.n++ }

type B struct{ C }

type A struct{ B }

func mkA() A { return A{} }

func main() {
	mkA().Inc()
}

// TypeCheckError:
// main/addressable_13e_err.gno:17:8: cannot call pointer method Inc on A

// Error:
// main/addressable_13e_err.gno:17:2-11: cannot call pointer method Inc on main.A
