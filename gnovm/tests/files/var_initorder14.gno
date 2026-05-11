package main

// Tests transitive dependency through chained method calls.
// A = T{}.MethodA(), MethodA calls MethodB, MethodB reads B.
// If findDependentNames only traces one level of method body,
// A's dependency on B (through MethodA→MethodB) might be missed,
// causing wrong init order.

type T struct{}

func (T) MethodA() int {
	t := T{}
	return t.MethodB()
}

func (T) MethodB() int {
	return B
}

var (
	A = T{}.MethodA()
	B = 42
)

func main() { println(A) }

// Output:
// 42
