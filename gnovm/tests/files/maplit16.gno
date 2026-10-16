// A typed nil slice, map or func key in a map literal is not a Go
// constant: duplicate-key detection must skip it rather than hash it.
package main

func never() {
	_ = map[any]int{[]int(nil): 1}
	_ = map[any]int{map[int]int(nil): 1}
	_ = map[any]int{(func())(nil): 1}
}

func main() {
	println("ok")
}

// Output:
// ok
