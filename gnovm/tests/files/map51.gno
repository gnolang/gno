package main

// Assigning to an existing key replaces the STORED key with the newly
// assigned one, so a later -0 assignment must win over an earlier 0.
// This is the invariant the displaced-key handoff in GetPointerForKey
// must preserve. See gnolang/gno#4114.

func main() {
	m := map[float64]string{}
	m[0] = "positive"
	m[-0.0] = "negative"
	println(len(m))
	println(m[0])
	println(m[-0.0])

	type key struct{ a int }
	n := map[key]int{}
	n[key{1}] = 10
	n[key{1}] = 20
	println(len(n))
	println(n[key{1}])
}

// Output:
// 1
// negative
// negative
// 1
// 20
