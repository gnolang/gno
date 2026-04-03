package main

// Tests that a dependency inside a function literal's body (single nesting
// level) is correctly attributed to the enclosing package-level ValueDecl.
// The inner `var local = B` is a local declaration; codaInitOrderDeps must
// walk past it and attribute the B reference to ValueDecl(A), not to the
// local ValueDecl.

var A = func() int {
	var (
		local = B
		_     = 123 // just to avoid `fmt` simplifying this into :=.
	)
	return local
}()

var B = 42

func main() {
	println(A)
}

// Output:
// 42
