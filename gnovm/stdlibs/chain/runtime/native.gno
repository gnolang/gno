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

// GetSessionInfo returns session metadata if the current transaction was
// signed by a session key. Returns isSession=false for master-key transactions.
//
// When isSession is true:
//   - pubKeyAddr is the session key's address (for per-session tracking)
//   - expiresAt is the unix timestamp when the session expires (0 = no expiry)
//   - allowPaths is the list of realm path prefixes this session can call;
//     nil or empty means unrestricted (any realm). Realms that want to enforce
//     session-specific limits should check isSession and apply their own
//     authorization logic.
func GetSessionInfo() (pubKeyAddr address, expiresAt int64, allowPaths []string, isSession bool) {
	addr, exp, paths, ok := getSessionInfo()
	return address(addr), exp, paths, ok
}

// Variations which don't use named types.
func originCaller() string
func getRealm(height int) (address string, pkgPath string)
func getSessionInfo() (pubKeyAddr string, expiresAt int64, allowPaths []string, isSession bool)
