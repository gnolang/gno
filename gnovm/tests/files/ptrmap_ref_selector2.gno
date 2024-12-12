// PKGPATH: gno.land/r/ptr_map
package ptr_map

type T struct {
	Field []int
}

func main() {
	t := T{Field: []int{1, 2, 3}}

	// AbsPath:  15bc39c5756bbda67fd0d56bc9e86b29d7b444fc:3:[0]:Field:(const (1 int))
	println(&t.Field[1])
}

// Output:
// &(2 int)
