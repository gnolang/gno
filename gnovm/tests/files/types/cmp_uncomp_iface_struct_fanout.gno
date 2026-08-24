package main

// The uncomparable field of Bad comes last: the comparability walk checks
// fields in order and stops at the first failure, so it must finish walking
// the entire fan-out subtree before the slice flips the verdict. Compared
// twice: the second comparison exercises the cached negative verdict.
// Depth here is semantic only — TestIsComparableMemoized in
// type_check_test.go guards the bounded-recursion property at depth 200.
type T0 struct{ x int }
type T1 struct{ a, b [0]T0 }
type T2 struct{ a, b [0]T1 }
type T3 struct{ a, b [0]T2 }

type Bad struct {
	t [0]T3
	s []int
}

func cmp(i, j interface{}) {
	defer func() { println("recover:", recover()) }()
	_ = i == j
}

func main() {
	var i, j interface{} = Bad{}, Bad{}
	cmp(i, j)
	cmp(i, j)
}

// Output:
// recover: runtime error: comparing uncomparable type main.Bad
// recover: runtime error: comparing uncomparable type main.Bad
