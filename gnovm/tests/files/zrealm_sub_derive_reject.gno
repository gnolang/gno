// PKGPATH: gno.land/r/demo/subderive
package subderive

import "chain"

// Drift-prevention: chain.DerivePkgSubAddr MUST reject exactly the same
// grammar that cur.Sub() rejects (its assertValidSubpath mirrors
// subRealmPathError byte-for-byte). If it didn't, an address could be
// derived off-chain that cur.Sub would refuse to mint — stranding any
// funds sent there. zrealm_sub2.gno only exercises the ACCEPTING path
// (DerivePkgSubAddr == Sub().Address() for a valid subpath); this file
// pins the REJECTING path, without which the mirror could silently
// drift and the equivalence claim would be untested on the failure side.
//
// It also asserts the derivation-safety check: a host containing
// "#" is rejected (forecloses nested synthesis and the
// run-path-extraction aliasing described in the ADR's "Derivation
// safety coupling").
func try(host, subpath string) {
	defer func() {
		println(recover())
	}()
	chain.DerivePkgSubAddr(host, subpath)
	println("UNEXPECTED: accepted", host, subpath)
}

func main() {
	const host = "gno.land/r/demo/subderive"

	// Invalid subpaths — same set the Sub() grammar rejects.
	try(host, "")     // empty
	try(host, "Dao")  // uppercase
	try(host, "a b")  // space
	try(host, "a:b")  // colon in subpath
	try(host, "dao/") // trailing slash
	try(host, "/dao") // leading slash
	try(host, "a//b") // empty segment
	try(host, "..")   // traversal
	try(host, "_x")   // edge punctuation

	// Invalid host: a synthesized "#"-bearing host must be rejected
	// (derivation-safety assert; subsumes the run-path assert since the
	// run-path shape is "#"-free).
	try(host+"#already", "x")

	// A run-path-shaped subpath is NOT special-cased away by DerivePkgSubAddr:
	// '/' is legal inside subpaths, so this is a normal (hashed) derivation,
	// NOT an address-extraction. It should be ACCEPTED and hash normally —
	// the safety rests on the host-colon-free assert, not on rejecting this.
	// (Kept as a positive control so a future change that starts extracting
	// addresses here is caught.)
	addr := chain.DerivePkgSubAddr(host, "e/g1abc/run")
	println("run-shaped subpath hashed (not extracted):",
		string(addr) != "g1abc" && addr == chain.PackageAddress(host+"#e/g1abc/run"))
}

// Output:
// Sub: subpath cannot be empty
// Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// Sub: subpath must be '/'-separated segments of [a-z0-9], with '_.-' allowed inside a segment
// Sub: receiver pkgpath is already synthesized or invalid
// run-shaped subpath hashed (not extracted): true
