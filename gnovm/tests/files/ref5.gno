package main

// ref of a nil interface{} variable, assign through pointer.

func main() {
	var i interface{}
	p := &i
	*p = 42
	println(i)
}

// Output:
// 42
