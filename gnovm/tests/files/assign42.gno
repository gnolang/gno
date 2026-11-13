package main

// An assignment before a shadowing declaration in the same block targets the
// outer name. Here that is a package-level func, which is not assignable; a
// name-keyed check stops at the shadow's slot and lets the write through.
func f() { println("original f") }

func caller() { f() }

func main() {
	caller()
	{
		f = func() { println("hijacked f") }
		f := 1
		_ = f
	}
	caller()
}

// Error:
// main/assign42.gno:13:3-39: cannot assign to func f

// TypeCheckError:
// main/assign42.gno:13:3: cannot assign to f (neither addressable nor a map index expression)
