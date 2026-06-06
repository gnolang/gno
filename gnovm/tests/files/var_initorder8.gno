package main

// Tests that a method call through a package-level receiver variable is
// correctly ordered. The SelectorExpr `obj.Val()` makes A syntactically
// depend on `obj`, and the struct literal `T{val: B}` makes obj depend on B.
// So the transitive chain B -> obj -> A is inferred without needing to
// inspect the method body.
// Init order: T (type, no deps), B (no deps), obj (deps T and B), A (deps obj).

type T struct{ val int }

func (t T) Val() int { return t.val }

var (
	B   = 42
	obj = T{val: B}
	A   = obj.Val()
)

func main() { println(A) }

// Output:
// 42
