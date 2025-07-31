// PKGPATH: gno.land/r/map_ref_key
package map_ref_key

import "strconv"

type myInt struct {
	num int
}

type stringMap map[int]myInt

var (
	accounts = stringMap{}
	tmpInt   = myInt{1}
)

func init() {
	accounts[1] = tmpInt
}

func Render(string) string {
	return strconv.Itoa(accounts[1].num)
}

func main() {
	println(Render("a"))
}

// Output:
// 1
