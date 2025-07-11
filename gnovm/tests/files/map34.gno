// PKGPATH: gno.land/r/mapkey
package mapkey

var (
	m = map[*int]int{}
)

func main() {
	i := 3
	tmpInt := &i
	m = map[*int]int{tmpInt: 1}
	println(m[tmpInt])
}

// Output:
// 1
