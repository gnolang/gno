// PKGPATH: gno.land/r/inline_arr_mutate
package inline_arr_mutate

// Test: init() mutation of inline array elements must persist to main().

var arr = [4]int{1, 2, 3, 4}

func init() {
	arr[2] = 999
}

func main(cur realm) {
	println(arr[2])
}

// Output:
// 999
