package main

// Outer compare enters via interface; dynamic type is a concrete array
// of a concrete uncomparable element type. viaInterface must propagate
// from the interface unwrap into the array element comparison.
//
// Go names the dynamic type carried by the interface header — the array
// type — not the inner map.
func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	var x interface{} = [1]map[int]int{nil}
	_ = x == x
}

// Output:
// recover: runtime error: comparing uncomparable type [1]map[int]int
