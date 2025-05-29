// PKGPATH: gno.land/r/map_test
package map_test

type myInt struct {
	num int
}

var M map[myInt]int = map[myInt]int{myInt{1}: 1}

func main() {
	println(M)
}

// Error:
// interface conversion: gnolang.Value is gnolang.RefValue, not *gnolang.StructValue
