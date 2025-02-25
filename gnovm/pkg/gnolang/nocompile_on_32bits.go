package gnolang

import "strconv"

func _() {
	// Restricting Gno to compile only on 64-bit architectures.
	// Please see https://github.com/gnolang/gno/issues/3288
	var x [1]struct{}
	_ = x[strconv.IntSize-64]
}
