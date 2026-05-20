// Package ownable (v1) provides the v3a-aligned ownership primitive
// for Gno realms. It is the successor to /p/nt/ownable/v0.
//
// # When to use v1 vs v0
//
//   - **New code**: prefer v1. The API is simpler (no `_ int, rlm realm`
//     parameter shape) and closes the omarsy forgery class structurally.
//   - **Existing code**: continue using v0 until the v3a migration
//     window closes. Both versions coexist during the transition.
//
// # Differences from v0
//
//   - Mutating methods take only data parameters; caller identity is
//     resolved by runtime.Caller() inside the method body.
//   - No `rlm.IsCurrent()` check. There is no rlm-typed value entering
//     these methods, so nothing to forge.
//   - New constructor NewWithCaller() captures the calling realm's
//     caller as the initial owner — equivalent to the v0 idiom of
//     using cur.Previous().Address() at init time.
//
// # Example
//
// In a realm that wants ownable behavior:
//
//	package vault
//
//	import "gno.land/p/nt/ownable/v1"
//
//	var owner = ownable.NewWithCaller()
//
//	func TransferOwnership(newOwner address) error {
//	    return owner.TransferOwnership(newOwner)
//	}
//
// The owner.TransferOwnership call internally queries runtime.Caller()
// and verifies it matches the recorded owner. Caller code passes only
// the new owner's address; no realm value flows across the boundary.
package ownable
