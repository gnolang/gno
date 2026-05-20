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

// Caller returns the realm of the frame that crossed into this realm —
// the v3a-aligned identity query. Equivalent to PreviousRealm under the
// current substrate (same walk machinery), but exposed as the canonical
// new-API name. Library helpers should use Caller() instead of taking a
// `rlm realm` parameter — that's the v2 pattern omarsy demonstrated is
// forgeable. The Realm value returned by Caller is a concrete struct
// constructed by the VM; user code has no path to supply a different
// Realm where Caller's return was used.
func Caller() Realm {
	return PreviousRealm()
}

// Self returns the realm currently executing this code. Equivalent to
// CurrentRealm, exposed under the v3a-aligned name.
func Self() Realm {
	return CurrentRealm()
}

// Origin returns the EOA realm that initiated the transaction (the bottom
// of the realm-transition stack). Equivalent to looking up the originating
// address but returned as a Realm value with empty pkgPath.
func Origin() Realm {
	return Realm{addr: address(originCaller()), pkgPath: ""}
}

// CallerN returns the realm n cross-realm transitions up the stack.
// CallerN(1) is equivalent to Caller(); CallerN(2) is "the realm that
// crossed into my caller's realm" — useful for ACL helpers that need
// to identify the principal two transitions up (e.g., the v2 pattern
// `rlm.Previous().Previous().Address()`).
//
// n=0 is equivalent to Self() (current realm). For n greater than the
// number of available transitions, behavior depends on the underlying
// getRealm walk — typically returns the EOA via OriginCaller for
// n = transitions+1, then panics for n beyond that.
func CallerN(n int) Realm {
	addr, path := getRealm(n)
	return Realm{address(addr), path}
}

// GetSessionInfo returns session metadata if the current transaction was
// signed by a session key. Returns isSession=false for master-key transactions.
//
// When isSession is true:
//   - pubKeyAddr is the session key's address (for per-session tracking)
//   - expiresAt is the unix timestamp when the session expires (0 = no expiry)
//   - allowPaths is the session's per-msg restriction list using the typed
//     grammar "*" or "<route>/<type>[:<path>]". Examples:
//     "*"                           — any msg (modulo always-denied)
//     "vm/exec:gno.land/r/foo/bar"  — MsgCall to that realm or sub-realms
//     "vm/exec"                     — any MsgCall
//     "vm/run"                      — any MsgRun
//     "bank/send"                   — any MsgSend
//     "bank/multisend"              — any MsgMultiSend
//     A msg passes if any entry matches its (route, type[, path]) AND it is
//     not in the always-denied list (auth/* and vm/add_package). AllowPaths
//     is required at create-time; an empty list is never produced. Realms
//     that want to enforce session-specific limits can also check isSession
//     and apply their own authorization logic.
func GetSessionInfo() (pubKeyAddr address, expiresAt int64, allowPaths []string, isSession bool) {
	addr, exp, paths, ok := getSessionInfo()
	return address(addr), exp, paths, ok
}

// Variations which don't use named types.
func originCaller() string
func getRealm(height int) (address string, pkgPath string)
func getSessionInfo() (pubKeyAddr string, expiresAt int64, allowPaths []string, isSession bool)
