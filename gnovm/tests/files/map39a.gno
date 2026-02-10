// PKGPATH: gno.land/r/map_ref_key
package map_ref_key

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

func main() {
	println(accounts)
}

// Output:
// (map{(1 int):(ref(0934e54b8f0575a8200841a7feee61eca343f4de:10) gno.land/r/map_ref_key.myInt)} gno.land/r/map_ref_key.stringMap)
