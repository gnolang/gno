package main

type empty interface{}
type boolean bool

func main() {
	m := map[int]int{1: 2}
	var ok any
	_, ok = m[0]
	println(ok.(bool))
	var value int
	value, ok = m[1]
	println(value, ok.(bool))

	var named empty
	_, named = m[0]
	println(named.(bool))
	_, named = m[1]
	println(named.(bool))

	var b boolean
	_, b = m[0]
	println(b)
	_, b = m[1]
	println(b)
}

// Output:
// false
// 2 true
// false
// true
// false
// true
