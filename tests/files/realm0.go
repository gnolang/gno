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
// u[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:0]=VI[object:OIDNONE:0#8B4A5E7CF86BD0F6828FE0332FF6E8415293B8FC&0]:
// - EI[:6DA88C34BA124C41F977DB66A4FC5C1A951708D2(#AE4B3280E56E2FAF83F414A6E3DABE9D5FBE1897)] // VI[numeric:02]
// - EI[nil]

// the above is showing that the realm's block (of 1 variable) changed.
