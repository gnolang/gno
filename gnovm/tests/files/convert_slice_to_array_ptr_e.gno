// spec: u0 := (*[0]byte)(make([]byte, 0))  // u0 != nil
package main

func main() {
	s := make([]int, 0)
	p := (*[0]int)(s)
	if p == nil {
		println("nil")
	} else {
		println("not nil")
	}
}

// Output:
// not nil
