// PKGPATH: gno.land/r/map_test
package map_test

type myInt struct {
	num int
}

var root interface{}

func main() {
	root = map[myInt]int{myInt{1}: 1} // mapLit
	println(root)
}

// Output:
// map{(struct{(1 int)} gno.land/r/map_test.myInt):(1 int)}
