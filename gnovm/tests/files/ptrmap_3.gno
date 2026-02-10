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
	j := *i
	return m[&j]
}

func init() {
	*i = 1
	AddToMap(5)
}

func main(cur realm,) {
	r := GetFromMap()
	println(r == 5)

	*i = 2
	r = GetFromMap()
	println(r == 5)
}

// Output:
// false
// false
