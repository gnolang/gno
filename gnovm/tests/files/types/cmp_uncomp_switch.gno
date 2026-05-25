package main

// switch on an interface-typed tag should panic when matching a case
// whose dynamic type is uncomparable (mirrors `x == case` semantics).
func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	var m map[int]int
	var x interface{} = m
	var y interface{} = m
	switch x {
	case y:
		println("matched")
	}
}

// Output:
// recover: comparing uncomparable type map[int]int
