// PKGPATH: gno.land/r/ptr_map
package ptr_map

var m = make(map[*int]string)

func foo(x, y int) {
	m[&x] = "a"
	m[&y] = "b"
	println(m[&x])
	println(m[&y])
}

func main() {
	foo(1, 1)
}

// Output:
// a
// b
