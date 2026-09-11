package main

// The shadowed and shadowing names are distinct slots, so a closure made
// before the shadowing declaration keeps capturing the outer one.
func main() {
	switch v := 1; v {
	case 1:
		f := func() int { return v }
		v := 99
		println("cap", f(), "shadow", v)
	}

	switch v := 1; v {
	case 1:
		v := 99
		g := func() int { return v }
		println("capshadow", g())
	}

	switch v := 1; v {
	case 1:
		v := 99
		println("inner", v)
		fallthrough
	case 2:
		println("outer", v)
	}
}

// Output:
// cap 1 shadow 99
// capshadow 99
// inner 99
// outer 1
