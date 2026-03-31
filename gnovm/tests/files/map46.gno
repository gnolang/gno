package main

func main() {
	// basic: insert then lookup via fresh &slice[i] expression
	bs := []byte("abc")
	m := map[*byte]int{}
	m[&bs[0]] = 1
	m[&bs[1]] = 2
	m[&bs[2]] = 3
	println(m[&bs[0]], m[&bs[1]], m[&bs[2]])

	// overwrite
	m[&bs[0]] = 99
	println(m[&bs[0]])

	// fixed [N]byte array
	var arr [3]byte
	m2 := map[*byte]int{}
	m2[&arr[0]] = 10
	println(m2[&arr[0]])
}

// Output:
// 1 2 3
// 99
// 10
