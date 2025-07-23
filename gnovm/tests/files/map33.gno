// PKGPATH: gno.land/r/mapkey
package mapkey

var i = 1
var m1 = map[*int]string{&i: "value"}

func main() {
	j := 1
	key := &j
	m1[key] = "value2"

	key0 := &i
	m1[key0] = "value3"

	println(m1[key])
	println(m1[key0])
}

// Output:
// value2
// value3
