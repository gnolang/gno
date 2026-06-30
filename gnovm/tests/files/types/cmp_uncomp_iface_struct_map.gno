package main

// Outer compare enters via interface; dynamic type is a struct whose
// field is a concrete map. viaInterface must propagate from the outer
// interface unwrap through the struct into the map field, so the
// uncomparable map field panics.
type S struct{ m map[int]int }

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	var x interface{} = S{}
	_ = x == x
}

// Output:
// recover: runtime error: comparing uncomparable type main.S
