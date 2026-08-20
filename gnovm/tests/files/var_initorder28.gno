package main

// Regression: codaInitOrderDeps must track method expressions on
// pointer types, e.g. (*T).PtrMethod (via the VPField path with
// *StarExpr unwrapping).

type T int

func (r *T) PtrMethod() int { return a }

var dummy = (*T).PtrMethod(new(T))

var a = identity(1)

func identity(a int) int { return a }

func main() {
	if dummy != 1 {
		panic("dummy != 1")
	}
	println(dummy)
}

// Output:
// 1
