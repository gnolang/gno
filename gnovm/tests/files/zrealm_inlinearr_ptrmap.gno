// PKGPATH: gno.land/r/inline_arr_ptrmap
package inline_arr_ptrmap

// Test: pointer to inline array element used as map key.
// Same pattern as ptrmap_24.gno but with arrays instead of structs.

var (
	arr = [4]int{10, 20, 30, 40}
	m   = map[*int]string{}
)

func init() {
	m[&arr[1]] = "second element"
}

func main(cur realm) {
	println(m[&arr[1]])
}

// Output:
// second element
