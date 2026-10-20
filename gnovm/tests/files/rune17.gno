package main

// Converting an integer constant to a float rounds once, from the exact
// value, to the nearest float (ties to even).

const (
	up   = '\x01'<<54 + 1<<30 + 1 // just above a float32 midpoint: rounds up
	tie  = '\x01'<<54 + 1<<30     // a float32 midpoint: ties to even (down)
	tie2 = '\x01'<<54 + 3<<30     // a float32 midpoint: ties to even (up)
	up64 = '\x01'<<83 + 1<<30 + 1 // just above a float64 midpoint
	big  = 1<<54 + 1<<30 + 1      // the same value, as an untyped int
)

func main() {
	println(int64(float32(up)), int64(float32(-up)))
	println(int64(float32(tie)), int64(float32(-tie)))
	println(int64(float32(tie2)), int64(float32(-tie2)))
	var a float32 = up
	var b float32 = -up
	println(int64(a), int64(b))
	println(int64(float32(big)), int64(float32(-big)))
	println(float64(up64) == float64('\x01'<<83+1<<31), float64(-up64) == -float64('\x01'<<83+1<<31))
}

// Output:
// 18014400656965632 -18014400656965632
// 18014398509481984 -18014398509481984
// 18014402804449280 -18014402804449280
// 18014400656965632 -18014400656965632
// 18014400656965632 -18014400656965632
// true true
