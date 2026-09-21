package main

// Names declared by the switch init statement stay live across fallthroughs,
// including into a trailing default clause.
func main() {
	switch v := 1; v {
	case 1:
		a, b := 10, 20
		println(a + b)
		fallthrough
	case 2:
		println("v is", v)
		fallthrough
	default:
		println("v still", v)
	}
}

// Output:
// 30
// v is 1
// v still 1
