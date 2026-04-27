// spec: only slice length matters; cap is irrelevant.
package main

func main() {
	s := make([]byte, 2, 4)
	p := (*[3]byte)(s)
	println(p)
}

// Error:
// runtime error: cannot convert slice with length 2 to pointer to array with length 3
