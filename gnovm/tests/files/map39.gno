// PKGPATH: gno.land/r/map_ref_key
package map_ref_key

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

func main() {
	println(accounts)
}

// Output:
// (map{(struct{(1 int)} gno.land/r/map_ref_key.myInt):(1 int)} gno.land/r/map_ref_key.stringMap)
