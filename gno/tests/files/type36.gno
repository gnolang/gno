package main

type Arr [2]int
type Ptr *Arr

func main() {
	var arr Arr
	arr[0] = 0
	arr[1] = 1

	p := Ptr(&arr)

	println(p[:])
}

// Output:
// slice[(0 int),(1 int)]
