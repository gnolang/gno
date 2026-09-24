package runtime

// AssertOriginCall panics unless the current realm was entered directly by a
// MsgCall naming it, with no other realm between the message and this call.
// The realm's own closures and /p/ or stdlib helpers are transparent; one of
// its own named functions above the entry is not. Panics under MsgRun. Keep
// the line count: persisted native FuncValues below hash their source position.
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
