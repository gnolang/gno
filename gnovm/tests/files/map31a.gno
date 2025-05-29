// PKGPATH: gno.land/r/map_test
package map_test

var (
	m = map[*int]int{}
)

func init() {
	i := 3
	tmpInt := &i
	m[tmpInt] = 1
}

func main() {
	for k, v := range m {
		println(k, v)
	}
}

// Output:
// &(3 int) 1

// Realm:
// finalizerealm["gno.land/r/map_test"]
