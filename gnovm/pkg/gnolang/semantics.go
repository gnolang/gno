package gnolang

// Per-package language semantics (WIP plumbing — see doc/adr, dormant).
//
// GOAL (not yet realized here): pin VM semantics to the gno language
// version a package declared at deploy time, so a later conformance fix
// (a "class-3" behavior change) applies only to packages that (re)deploy
// under the new version — deployed code keeps the semantics it was
// written and audited against. This is the gno analogue of Solidity's
// pragma / Go's -lang: the mechanism that lets the VM evolve without
// silently re-meaning already-deployed contracts.
//
// WHAT THIS FILE ESTABLISHES: the seam only. A Semantics value is
// resolved from a package's language version and is the single place
// future version-conditional behavior will branch. Today exactly one
// version (GnoVerLatest) is registered and every field is current
// behavior, so consulting Semantics changes nothing — it is dormant.
//
// SOURCE OF THE VERSION — NO NEW PERSISTED FIELD NEEDED. The gno
// language version is ALREADY persisted with every package: addpkg
// writes gnomod.toml into the stored MemPackage (vm keeper), and the
// store can read it back at load via GetMemFile(path, "gnomod.toml").
// So the real design is to SOURCE PackageNode.LangVersion from that
// existing stored gnomod.toml at package load — not to add a new
// serialized field. This keeps the consensus-visible stored form
// unchanged (the version is already inside the hashed MemPackage), the
// same way EVM freezes semantics into the immutable deployed artifact.
// (LangVersion below is a WIP non-serialized default until that wiring
// lands; it is dormant, so the default is a no-op today.)
//
// WHAT IS DELIBERATELY NOT DONE YET (the real follow-up work):
//   - LOAD-TIME WIRING: populate LangVersion by reading the stored
//     gnomod.toml when a package is loaded, instead of defaulting.
//   - STOP FORCING LATEST AT DEPLOY: addpkg currently overwrites
//     gm.Gno = GnoVerLatest (vm keeper). That is harmless with one
//     version but DEFEATS future pinning — an old package must keep its
//     deploy-time version in its stored gnomod so it still runs under
//     old semantics after the VM advances. Must record the real
//     deploy-time version instead.
//   - PER-FRAME DISPATCH across a mixed call stack (v1 package calling a
//     v2 package): Semantics must be scoped per executing package, with
//     defined conversion rules at the boundary. NOTE: only semantics
//     changes that are INVISIBLE across the import/call boundary are
//     pinnable this way; a change observable at the boundary (shared
//     type identity / value representation / serialization) is
//     shared-substrate and must go through a re-genesis epoch, not a pin.
//   - Path versioning (xxx/yy/v2) is ORTHOGONAL: it lets an author ship
//     new CODE that dependents opt into by import path. It does NOT pin
//     semantics — an unchanged deployed package at its old path still
//     runs on whatever VM is live, so it cannot protect that package
//     from a VM behavior change. Pinning is what does that.
//   - GAS is out of scope: the gas schedule is a chain-wide resource,
//     height-versioned, never per-package (see 3b).

// Semantics captures the version-conditional behavior switches the VM
// consults. It is derived from a package's declared language version.
// All fields describe CURRENT behavior; new fields get a default equal
// to today's behavior so that omitting the version is a no-op.
type Semantics struct {
	// Version is the gno language version these semantics correspond to
	// (e.g. GnoVerLatest). Informational; feature switches are the
	// operative part.
	Version string

	// (No feature switches yet. When the first class-3 conformance fix
	// needs to be version-gated, add a bool here — false = pre-fix
	// behavior for older-version packages, true = fixed behavior for
	// packages on the version that introduced the fix. The registry in
	// SemanticsForVersion is where each version sets them.)
}

// semanticsByVersion is the registry of supported language versions.
// A map (not a switch) so a second version — and, in tests, a synthetic
// one — is a single entry. Production ships exactly one entry until the
// persistence follow-up lands (see file header).
var semanticsByVersion = map[string]Semantics{
	GnoVerLatest: {Version: GnoVerLatest},
}

// SemanticsForVersion returns the Semantics for a gno language version,
// and whether that version is supported. An unknown version returns
// ok=false so callers can reject rather than silently run under an
// undefined semantics.
func SemanticsForVersion(version string) (Semantics, bool) {
	s, ok := semanticsByVersion[version]
	return s, ok
}

// registerSemanticsForTest adds a synthetic version to the registry and
// returns a cleanup func. Test-only: it exists so the per-package
// dispatch can be exercised (a real second version can't be committed
// before the persistence follow-up). Not exported — callers live in the
// same package's _test.go.
func registerSemanticsForTest(s Semantics) (cleanup func()) {
	if _, exists := semanticsByVersion[s.Version]; exists {
		panic("registerSemanticsForTest: version already registered: " + s.Version)
	}
	semanticsByVersion[s.Version] = s
	return func() { delete(semanticsByVersion, s.Version) }
}

// Semantics returns the resolved semantics for this package. It never
// fails: an unset or unrecognized LangVersion falls back to GnoVerLatest
// (the only registered version), keeping the seam dormant until a second
// version and its persistence are wired.
func (pn *PackageNode) Semantics() Semantics {
	v := pn.LangVersion
	if s, ok := SemanticsForVersion(v); ok {
		return s
	}
	s, _ := SemanticsForVersion(GnoVerLatest)
	return s
}
