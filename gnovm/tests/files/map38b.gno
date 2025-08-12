// PKGPATH: gno.land/r/map_ref_key
package map_ref_key

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

func main() {
	println(accounts)
}

// Output:
// (map{(struct{(1 int),(struct{(2 int)} gno.land/r/map_ref_key.MyInt2)} gno.land/r/map_ref_key.myInt):(1 int)} gno.land/r/map_ref_key.stringMap)
