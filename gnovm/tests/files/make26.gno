package main

// A trailing element after a near-max key overflows the measured length. The
// key MaxInt64-1 sets the running index to MaxInt64, then the unkeyed element
// `2` takes index MaxInt64 and pushes the implied length to MaxInt64+1, which
// wraps int64 negative. Like make23, Go rejects the overflowing element as an
// out-of-bounds index; gno must reject it at compile time, not silently accept
// a negative array length.

func main() {
	x := [...]int{9223372036854775806: 1, 2}
	_ = x
}

// Error:
// main/make26.gno:11:7-42: array index 9223372036854775807 out of bounds [0:0]
