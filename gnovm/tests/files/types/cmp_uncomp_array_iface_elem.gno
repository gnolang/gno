package main

// Concrete array compare (outer not via interface) where the element
// type is interface{} — viaInterface must propagate into the elements
// so the uncomparable dynamic type panics.
func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	var m map[int]int
	a1 := [1]interface{}{m}
	a2 := [1]interface{}{m}
	_ = a1 == a2
}

// Output:
// recover: runtime error: comparing uncomparable type map[int]int
