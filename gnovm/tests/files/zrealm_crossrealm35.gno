// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

type XYZ struct{ name string }

var S []*XYZ

func init() {
	S = append(S, &XYZ{"1"})
}

func main() {
	println(S[0].name)
}

// Output:
// 1
