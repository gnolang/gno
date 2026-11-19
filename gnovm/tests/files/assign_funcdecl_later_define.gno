package main

func f() { println("f") }

// A local f declared later in the block does not make the package func f
// assignable before it.
func main() {
	f = func() { println("hijacked") }
	f()
	f := 1
	_ = f
}

// Error:
// main/assign_funcdecl_later_define.gno:8:2-36: cannot assign to func f

// TypeCheckError:
// main/assign_funcdecl_later_define.gno:8:2: cannot assign to f (neither addressable nor a map index expression)
