// PKGPATH: gno.land/r/ptr_slice
package ptr_slice

var (
	li = []*int{}
	i  = new(int)
)

func init() {
	li = append(li, i)
}

func main() {
	println(li[0] == i)
}

// Output:
// true
