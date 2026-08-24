package main

type Record struct {
	Value int
	Label string
}

func main() {
	var a [2]Record
	a[0] = Record{Value: 100, Label: "original-0"}
	a[1] = Record{Value: 200, Label: "original-1"}

	// Array assignment must produce a deep copy (Go value semantics).
	// Mutating b must never affect a.
	b := a
	b[0].Value = 999
	b[0].Label = "modified-0"
	b[1].Value = 888
	b[1].Label = "modified-1"

	println(a[0].Label, a[0].Value)
	println(a[1].Label, a[1].Value)
	println(b[0].Label, b[0].Value)
	println(b[1].Label, b[1].Value)

	// Also test pass-by-value: modifying a function's array parameter
	// must not affect the caller's copy.
	modify(a)
	println(a[0].Label, a[0].Value)
}

func modify(arr [2]Record) {
	arr[0].Value = 777
	arr[0].Label = "param-modified"
}

// Output:
// original-0 100
// original-1 200
// modified-0 999
// modified-1 888
// original-0 100
