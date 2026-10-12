// PKGPATH: gno.land/r/b64test
package b64test

// A stdlib method (base64.Encode) writing into a CALLER-supplied, real
// /r/-stamped out-param buffer now works: borrow rule #2 is skipped for
// stdlib-stamped receivers, so m.Realm stays /r/b64test and the write to
// the caller's own buffer passes the readonly gate.

import "encoding/base64"

var dst [4]byte // /r/b64test-stamped, real after init; EncodedLen("hi")==4

func main(cur realm) {
	base64.StdEncoding.Encode(dst[:], []byte("hi"))
	println(string(dst[:]))
}

// Output:
// aGk=
