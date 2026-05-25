package main

// Non-nil map via interface{} must also panic (the dynamic type is
// uncomparable regardless of value).
func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	m := map[int]int{1: 1}
	var x interface{} = m
	_ = x == x
}

// Output:
// recover: runtime error: comparing uncomparable type map[int]int
