package main

// Interface-keyed maps must keep the TypeID prefix: these values all
// serialize to identical value bytes and are separated only by their type.

func main() {
	m := map[any]string{}
	m[int(1)] = "int"
	m[int64(1)] = "int64"
	m[uint64(1)] = "uint64"
	m[int32(1)] = "int32"
	m[uint32(1)] = "uint32"
	println(len(m))
	println(m[int(1)])
	println(m[int64(1)])
	println(m[uint64(1)])
	println(m[int32(1)])
	println(m[uint32(1)])

	// The map's VALUE type being an interface must not affect key handling.
	n := map[int]any{}
	n[1] = "one"
	n[2] = "two"
	println(len(n))
	println(n[1])
}

// Output:
// 5
// int
// int64
// uint64
// int32
// uint32
// 2
// one
