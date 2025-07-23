// PKGPATH: gno.land/r/map_ref_key
package map_ref_key

import (
	"strconv"
)

type myInt struct {
	num int
}

type stringMap map[myInt]int

var (
	accounts = stringMap{}
	tmpInt   = myInt{1}
)

func init() {
	accounts[tmpInt] = 1
}

func Render(string) string {
	return "[Break the map!](maperror$help&func=Register)\n\nvalue: " + strconv.Itoa(accounts[tmpInt])
}

func main() {
	println(Render("a"))
}

// Output:
// [Break the map!](maperror$help&func=Register)
//
// value: 1
