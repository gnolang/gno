package main

// slice -> *array
func main() {
	s0 := "hello"
	s1 := []byte(s0)
	arrPtr1 := (*[5]byte)(s1)
	arr := [5]byte(s1)
	arrPtr2 := &arr

	println(arrPtr1 == arrPtr2)
}

// Output:
// false
