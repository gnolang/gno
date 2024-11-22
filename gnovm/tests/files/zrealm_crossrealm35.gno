// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

type XYZ struct{ name string }

var S []*XYZ

func init() {
	S = append(S, &XYZ{"1"})
}

func main() {
	a := S[0].name
	println(a)
}

// Output:
// 1
