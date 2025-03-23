package std

func AssertOriginCall() // injected

// CallerAt calls callerAt, which we overwrite
func callerAt(n int) string

// native bindings
func getRealm(height int) (address string, pkgPath string)
func isRealm(pkgPath string) bool
func originPkgAddr() string
