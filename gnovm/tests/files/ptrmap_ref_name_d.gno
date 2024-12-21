// PKGPATH: gno.land/r/ptr_map
package ptr_map

var arr = [2]int{1, 2} // Take the address of the array variable
var m = map[*[2]int]string{}

func init() {
	m[&arr] = "ok"
}

func main() {

	println(m[&arr]) // Output: example
}

// Output:
// ok
