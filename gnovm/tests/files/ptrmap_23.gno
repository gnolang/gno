// PKGPATH: gno.land/r/ptr_map
package ptr_map

var arr = [2]int{1, 2}
var m = map[*[2]int]string{}

func init() {
	m[&arr] = "ok"
}

func main(cur realm,) {
	println(m[&arr]) // Output: example
}

// Output:
// ok
