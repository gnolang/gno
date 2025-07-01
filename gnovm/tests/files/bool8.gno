package main

// results from comparisons should not be untyped bools

var a any = true

func main() {
	buf := "hello="
	isEqual(a, (buf[len(buf)-1] == '='))
}

func isEqual(v1, v2 any) {
	println("v1 == v2", v1 == v2)
}

// Output:
// v1 == v2 true
