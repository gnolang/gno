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

// ObjectID returns the identity the VM minted for the object v refers to,
// formatted as "<realm-id>:<clock-tick>", or "" when v names no object that
// has one.
//
// It exists because an identifier a realm chooses is an identifier a realm can
// choose twice (gnolang/gno#6026). This one is not chosen: the VM assigns it
// while persisting the object, from the owning realm's storage clock, so two
// objects a realm created under one name still differ here.
//
// Only a pointer to a standalone object answers. A pointer into a struct field
// or an array element resolves to its container, and a slice to its backing
// array, so lending either the container's identity would have every sibling
// report the same one; both get "" instead. The set of shapes that answer is
// meant to grow and only to grow: widening it later is compatible, taking an
// answer back is not.
//
// "" covers two situations a caller should tell apart, and the right way to
// tell them apart is to know what was passed:
//
//   - v is a pointer to an object the running call created. Identity is minted
//     when the realm frame returns, so ask again from a later call. Objects
//     created in init are stamped by the time any call runs.
//   - v has nothing to wait for: a scalar, a nil pointer, a value owned by a
//     stdlib or a p/ package.
//
// It never panics and never advances a realm's clock, so repeated calls agree
// and reading an identity has no effect on the object's lifecycle.
//
// The string is an identifier, not an account: it is not an address and
// nothing can be sent to it.
func ObjectID(v any) string // injected
