package main

// Like assign42.gno, but the shadow is a type declared in a nested block.
func f() { println("original f") }

func caller() { f() }

func main() {
	caller()
	{
		f = func() { println("hijacked f") }
		type f struct{ n int }
		_ = f{1}
	}
	caller()
}

// Error:
// main/assign46.gno:11:3-39: cannot assign to func f

// TypeCheckError:
// main/assign46.gno:11:3: cannot assign to f (neither addressable nor a map index expression)
