// PKGPATH: gno.land/r/map_test
package map_test

var (
	m = map[*int]int{}
)

func main() {
	i := 3
	tmpInt := &i
	m[tmpInt] = 1
}
