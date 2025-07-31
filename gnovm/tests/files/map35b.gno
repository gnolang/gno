// PKGPATH: gno.land/r/map_ref_key
package map_ref_key

import "strconv"

type myInt struct {
	num    int
	myint2 MyInt2
}

type MyInt2 struct {
	num2 int
}

type stringMap map[myInt]int

var (
	accounts = stringMap{}
	tmpInt   = myInt{num: 1, myint2: MyInt2{num2: 2}}
)

func init() {
	accounts[tmpInt] = 1
}

func Render(string) string {
	return "[Break the map!](maperror$help&func=Register)\n\nmap len: " + strconv.Itoa(len(accounts))
}

func main() {
	println(Render("a"))
}

// Output:
// [Break the map!](maperror$help&func=Register)
//
// map len: 1
