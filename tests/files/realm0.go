// PKGPATH: gno.land/r/test
package test

var root interface{}

func main() {
	println(root)
	root = 1
	println(root)
}

// Output:
// nil
// 1

// Realm:
// u[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:0]=VI[object:OIDNONE:0#905B893E53B6333B757E3F3221C57137AF6481EF@1&0]:
// - EI[nil]
// - EI[:6DA88C34BA124C41F977DB66A4FC5C1A951708D2(#AE4B3280E56E2FAF83F414A6E3DABE9D5FBE1897)] // VI[numeric:02]

// The above is showing that the realm's block (of 1 variable) changed.  The
// first EI (element image) is for the "main" function, which appears first
// because function declarations are defined in a file before vars, and is nil
// because functions don't have any image value.
