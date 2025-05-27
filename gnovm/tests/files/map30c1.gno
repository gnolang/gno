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
