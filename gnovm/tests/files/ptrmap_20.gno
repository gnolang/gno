// PKGPATH: gno.land/r/ptr_map
package ptr_map

var (
	m = map[**int]int{}
	i = new(int)
)

func AddToMap(value int) {
	m[&i] = value
}

func GetFromMap() int {
	return m[&i]
}

func init() {
	*i = 1
	AddToMap(5)
}

func main(cur realm,) {
	r := GetFromMap()
	println(r == 5)

	var j = 0
	j1 := &j
	println(m[&j1])

	*i = 2
	r = GetFromMap()
	println(r == 5)
}

// Output:
// true
// 0
// true
