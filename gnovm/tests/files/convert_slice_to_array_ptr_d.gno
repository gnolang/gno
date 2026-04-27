// spec: t1 := (*[0]string)(t) where t is nil → t1 == nil
package main

func main() {
	var s []int
	p := (*[0]int)(s)
	if p == nil {
		println("nil")
	} else {
		println("not nil")
	}
}

// Output:
// nil
