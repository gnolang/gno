package main

type empty interface{}
type boolean bool

func main() {
	var x any = 2
	var ok any
	_, ok = x.(string)
	println(ok.(bool))
	var value int
	value, ok = x.(int)
	println(value, ok.(bool))

	var named empty
	_, named = x.(string)
	println(named.(bool))
	_, named = x.(int)
	println(named.(bool))

	var b boolean
	_, b = x.(string)
	println(b)
	_, b = x.(int)
	println(b)
}

// Output:
// false
// 2 true
// false
// true
// false
// true
