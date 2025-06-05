// PKGPATH: gno.land/r/ptr_map
package ptr_map

var (
	m       = map[*int]int{}
	a, b, c = 1, 2, 3
	s       = []*int{&a, &b, &c}
)

func AddToMap(value int) {
	m[&*s[0]] = value
}

func GetFromMap() int {
	return m[&*s[0]]
}

func init() {
	AddToMap(5)
}

func main(cur realm,) {
	r := GetFromMap()
	println(r == 5)

	a = 0

	r = GetFromMap()
	println(r == 5)
}

// Output:
// true
// true
