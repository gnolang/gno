// PKGPATH: gno.land/r/map_test
package map_test

type myInt struct {
	num int
}

var M map[myInt]int = map[myInt]int{myInt{1}: 1} // mapLit

func main() {
	println(M)
}

// Output:
// map{(struct{(1 int)} gno.land/r/map_test.myInt):(1 int)}
