// PKGPATH: gno.land/r/ptr_map
package ptr_map

var (
	m = map[*int]int{}
	i = new(int)
)

func AddToMap(value int) {
	m[i] = value
}

func GetFromMap() int {
	return m[i]
}

func init() {
	*i = 1
	AddToMap(5)
}

// ----above is initialized and persisted before main is executed.

func main(cur realm,) {
	r := GetFromMap()
	println(r == 5)

	*i = 2 // this changes TV, also Base of a pointer value
	r = GetFromMap()
	println(r == 5)
}

// Output:
// true
// true
