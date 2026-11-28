// Alias-embed naming through the RUNTIME struct-construction path
// (doOpStructType): this combination — local type decl + struct assertion
// target + anonymous struct literal in a closure — is pinned (by
// instrumentation) to build struct{ Int } via the machine op rather than
// preprocess-only, and x == w observes that the runtime-built type keeps the
// alias spelling: it equals the preprocess-built struct{ Int } and stays
// distinct from struct{ int }. Matches Go. (TestDoOpStructType_EmbedNames
// drives the op directly for the same rule.)
package main

type Int = int

func main() {
	type S struct{ Int }
	var v interface{} = S{}
	_, ok := v.(struct{ Int })
	println(ok)
	f := func() interface{} { return struct{ Int }{} }
	x := f()
	println(x == v)
	var w interface{} = struct{ Int }{}
	println(x == w)
	var u interface{} = struct{ int }{}
	println(x == u)
}

// Output:
// false
// false
// true
// false
