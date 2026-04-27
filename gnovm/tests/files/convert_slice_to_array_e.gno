// spec: t0 := [0]string(t)  // ok for nil slice t
package main

func main() {
	var t []string
	a := [0]string(t)
	println(len(a))
}

// Output:
// 0
