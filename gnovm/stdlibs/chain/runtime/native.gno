package runtime

// AssertOriginCall panics if the calling method is not invoked via a direct
// MsgCall. It panics for other cases, like if the calling method
// is invoked by another method (even from the same realm or package).
// It also panic every time when the transaction is broadcasted via
// MsgRun.
func AssertOriginCall()   // injected
func ChainID() string     // injected
func ChainDomain() string // injected
func ChainHeight() int64  // injected

func OriginCaller() address {
	return address(originCaller())
}

func CurrentRealm() Realm {
	addr, path := getRealm(0)
	return Realm{address(addr), path}
}

func PreviousRealm() Realm {
	addr, path := getRealm(1)
	return Realm{address(addr), path}
}

// Variations which don't use named types.
func originCaller() string
func getRealm(height int) (address string, pkgPath string)
