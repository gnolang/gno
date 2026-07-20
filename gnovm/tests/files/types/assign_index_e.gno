// v, ok = on a declared-map composite literal index: valid, used to crash
// with "should not happen".
package main

type M map[string]int

func main() {
	var v int
	var ok bool
	v, ok = M{"a": 1}["a"]
	println(v, ok)
}

// Output:
// 1 true
