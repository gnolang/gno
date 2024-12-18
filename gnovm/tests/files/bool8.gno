package main

// results from comparisons should not be untyped bools

var a interface{} = true

func main() {
	buf := "hello="
	isEqual(a, (buf[len(buf)-1] == '='))
}

func isEqual(v1, v2 interface{}) {
	println("v1 == v2", v1 == v2)
}

// Output:
// v1 == v2 true
