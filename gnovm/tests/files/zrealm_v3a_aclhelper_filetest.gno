// PKGPATH: gno.land/r/v3aaclhost
package v3aaclhost

import (
	"chain/runtime"

	"gno.land/p/demo/tests/v3aclhelper"
)

// Demonstrates the v3a-aligned ACL pattern: a /p/ helper identifies
// its caller via runtime.Caller() instead of taking a `rlm realm`
// parameter. The caller passes only data (the expected address);
// no realm value flows through, so the omarsy forgery class has no
// surface here.
//
// Stack at v3aclhelper.CheckCaller:
//
//	EOA → /r/v3aaclhost.main → /p/v3aclhelper.CheckCaller
//	      m.Realm shifts to    /p/ borrows caller realm
//	      /r/v3aaclhost        (m.Realm = /r/v3aaclhost)
//
// Inside CheckCaller, runtime.Caller() walks past the same-realm
// /p/ frame (which inherited /r/v3aaclhost's realm) and lands on
// the EOA's MsgCall frame, returning the EOA realm.

func main(cur realm) {
	eoaAddr := runtime.PreviousRealm().Address()

	// runtime.Caller() inside /p/v3aclhelper sees the EOA.
	println("matches eoa:", v3aclhelper.CheckCaller(eoaAddr))

	// CallerPath returns the EOA's pkgPath, which is empty.
	println("path empty:", v3aclhelper.CallerPath() == "")

	// CallerIsUser is true since EOA is the caller.
	println("is user:", v3aclhelper.CallerIsUser())
}

// Output:
// matches eoa: true
// path empty: true
// is user: true
