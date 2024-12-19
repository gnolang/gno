// PKGPATH: gno.land/r/ptr_map
package ptr_map

var (
	m = map[*int]int{}
	i = 0
)

func AddToMap(value int) {
	m[&i] = value
}

func GetFromMap() int {
	i := 0 // same name and value path(3, 1), but different `origin`
	{
		{
			return m[&i]
		}
	}
}

func init() {
	AddToMap(5)
}

func main() {
	r := GetFromMap()
	println(r == 5)

	i = 2
	r = GetFromMap()
	println(r == 5)
}

// Output:
// false
// false
