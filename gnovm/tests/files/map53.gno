package main

// Concrete key types across the shapes that reach each ComputeMapKey branch:
// a declared type over a primitive (the map[address]T shape), a pointer key
// including nil, an array key, and a struct key whose field is an interface
// (so the nested omitTypes rule still applies inside an omitted-prefix key).

type Address string

type boxed struct {
	tag  string
	body any
}

func main() {
	addrs := map[Address]int{}
	addrs["g1abc"] = 1
	addrs["g1def"] = 2
	println(len(addrs))
	println(addrs["g1abc"])
	delete(addrs, "g1abc")
	println(len(addrs))
	_, ok := addrs["g1abc"]
	println(ok)

	type node struct{ n int }
	a, b := &node{1}, &node{2}
	ptrs := map[*node]string{}
	ptrs[a] = "a"
	ptrs[b] = "b"
	ptrs[nil] = "nil"
	println(len(ptrs))
	println(ptrs[a])
	println(ptrs[nil])

	arrs := map[[2]int]string{}
	arrs[[2]int{1, 2}] = "one-two"
	arrs[[2]int{2, 1}] = "two-one"
	println(len(arrs))
	println(arrs[[2]int{1, 2}])

	// Struct key with an interface field: the outer prefix is omitted, the
	// inner field's prefix is not, so int64(1) and int32(1) stay distinct.
	boxes := map[boxed]string{}
	boxes[boxed{"x", int64(1)}] = "i64"
	boxes[boxed{"x", int32(1)}] = "i32"
	println(len(boxes))
	println(boxes[boxed{"x", int64(1)}])
}

// Output:
// 2
// 1
// 1
// false
// 3
// a
// nil
// 2
// one-two
// 2
// i64
