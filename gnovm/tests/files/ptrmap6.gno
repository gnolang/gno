// PKGPATH: gno.land/r/ptr_map
package ptr_map

import "fmt"

type MyStruct struct {
	Name string
}

var (
	m    = map[[1]*MyStruct]string{}
	i1   = &MyStruct{Name: "alice"}
	key1 = [1]*MyStruct{i1}
)

func init() {
	m[key1] = "first key"
}

func main() {
	fmt.Println("Value for key1:", m[key1])
}

// Output:
// Value for key1: first key
