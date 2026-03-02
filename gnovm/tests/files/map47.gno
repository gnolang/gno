package main

func main() {
	// basic: insert then lookup via fresh &slice[i] expression
	bs := []byte("abc")
	bs = bs[:len(bs):len(bs)]
	m := map[*byte]int{}
	m[&bs[0]] = 1
	m[&bs[1]] = 2
	m[&bs[2]] = 3

	println(m[&bs[0]], m[&bs[1]], m[&bs[2]])
	bs = append(bs, 'd') // reallocate bs to force new addresses for &bs[i]
	println(m[&bs[0]], m[&bs[1]], m[&bs[2]])
}

// Output:
// 1 2 3
// 0 0 0
