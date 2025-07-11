// PKGPATH: gno.land/r/mapkey
package mapkey

type myInt struct {
	num int
}

var root interface{}

func main() {
	// map is new real, its children is handled via incRefCreatedDescendants.
	root = map[myInt]int{myInt{1}: 1}
	println(root)
}

// Output:
// map{(struct{(1 int)} gno.land/r/mapkey.myInt):(1 int)}
