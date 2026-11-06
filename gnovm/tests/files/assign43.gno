package main

// A range clause assigning into a package-level func name is rejected like a
// plain assignment: both reach assertValidAssignLhs.
func f() { println("original f") }

func main() {
	for _, f = range []func(){func() { println("hijacked f") }} {
	}
	f()
}

// Error:
// main/assign43.gno:8:2-9:3: cannot assign to func f

// TypeCheckError:
// main/assign43.gno:8:9: cannot assign to f (neither addressable nor a map index expression)
