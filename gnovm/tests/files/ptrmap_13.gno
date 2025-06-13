// PKGPATH: gno.land/r/ptr_map
package ptr_map

type MyStruct struct {
	Index int
}

var m = make(map[*int]string)
var a, b = 1, 2
var s = []*int{&a, &b}
var myStruct = MyStruct{Index: 0}

func init() {
	m[s[myStruct.Index]] = "a"
}

func main(cur realm,) {
	println(m[s[myStruct.Index]])
}

// Output:
// a
