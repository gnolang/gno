package main

// Mutual type-decl recursion: T2's RHS resolves while T1 is still unsealed.

type T1 struct {
	Next *T2
}

type T2 T1

func main() {
	var t T1
	t.Next = &T2{}
	println("ok")
}

// Output:
// ok
