// PKGPATH: gno.land/r/ptr_map
package ptr_map

type S struct {
	i int
}

var m = make(map[*S]string)          // Initialize the map
var sArr = []*S{&S{1}, &S{2}, &S{3}} // Use a slice of pointers

func init() {
	m[sArr[1]] = "a"
}

func main(cur realm,) {
	// Create a new slice without reallocating memory for existing elements
	newArr := append(sArr[:1], sArr[2:]...) // same base array

	// Compare pointers directly
	println(sArr[1] == newArr[1])
	println(m[sArr[1]] == m[newArr[1]])
	println(m[newArr[1]] == "")
}

// Output:
// true
// true
// true
