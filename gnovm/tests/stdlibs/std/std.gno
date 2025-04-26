package std

func AssertOriginCall() // injected

/* See comment in stdlibs/std/native.gno
// CallerAt calls callerAt, which we overwrite
func callerAt(n int) string
*/

// native bindings
func getRealm(height int) (address string, pkgPath string)
func isRealm(pkgPath string) bool
