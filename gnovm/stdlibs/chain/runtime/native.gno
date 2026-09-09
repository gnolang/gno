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

// ObjectID returns the address derived from v's own VM object ID: a stable,
// realm-scoped name for that object, which the realm chooses no part of.
//
// An object is stamped the first time it is persisted, which happens when a
// realm frame returns, so a value the running call created reads "" until then.
// Afterwards the address never changes. A pointer into a struct field or an
// array element, or a slice, reads "" as well: only a value with its own heap
// item has an address. It panics if v carries no object, such as a plain int
// or a nil pointer.
//
// The address has no signing key. Nothing on chain can spend coins sent to it,
// so use it as an identifier, not as a payable account.
func ObjectID(v interface{}) string { return objectID(v) }

func objectID(v interface{}) string // injected

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
func getSessionInfo() (pubKeyAddr string, expiresAt int64, allowPaths []string, isSession bool)
