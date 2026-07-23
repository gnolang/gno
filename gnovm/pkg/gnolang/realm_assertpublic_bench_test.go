package gnolang

import (
	"strconv"
	"testing"
)

// These benchmarks isolate the cost of assertTypeIsPublic itself — the
// type-graph walk on a miss vs. the store-cache lookup on a hit — which
// is exactly what typeHasPrivateDep's cache removes. They deliberately
// build the type ONCE, outside the timed loop, and do NOT include the
// per-commit cost of materializing the type object (GetType / amino
// decode): that cost is unchanged by this cache and would only swamp the
// signal.
//
// Object identity is irrelevant here: the cache keys on TypeID, so a
// fresh object reloaded in a later transaction hits exactly as a reused
// object does. That cross-transaction, fresh-object survival — the
// property the previous per-object design lacked — is verified separately
// by TestTypeHasPrivateDep_CacheSurvivesTransaction, not re-measured here.

// registerPublicPkgs registers numPkgs public leaf packages plus a public
// root package under pkgPrefix, and returns their paths. Registration is
// separate from type construction so a test can build many fresh,
// same-TypeID type objects against one already-registered package set.
func registerPublicPkgs(store Store, pkgPrefix string, numPkgs int) (pkgPaths []string, rootPath string) {
	pkgPaths = make([]string, numPkgs)
	for i := range pkgPaths {
		pkgPaths[i] = pkgPrefix + "/pkg" + strconv.Itoa(i)
		store.SetCachePackage(&PackageValue{PkgPath: pkgPaths[i], Private: false})
	}
	rootPath = pkgPrefix + "/root"
	store.SetCachePackage(&PackageValue{PkgPath: rootPath, Private: false})
	return pkgPaths, rootPath
}

// buildPublicType builds a fresh StructType with numFields fields, each a
// distinct nested StructType spread across the given leaf packages —
// representative of a mid-sized realm struct embedding a few imported
// types. Called repeatedly it yields distinct objects that all share one
// TypeID (structurally identical), modelling GetType across commits.
func buildPublicType(pkgPaths []string, rootPath string, numFields int) *StructType {
	fields := make([]FieldType, numFields)
	for i := range fields {
		nested := &StructType{
			PkgPath: pkgPaths[i%len(pkgPaths)],
			Fields: []FieldType{
				{Name: "A", Type: IntType},
				{Name: "B", Type: StringType},
				{Name: "C", Type: BoolType},
			},
		}
		fields[i] = FieldType{Name: Name("F" + strconv.Itoa(i)), Type: nested}
	}
	return &StructType{PkgPath: rootPath, Fields: fields}
}

// buildSelfReferentialType builds a fresh DeclaredType shaped like
// gno.land/p/nt/avl's Node (leftNode/rightNode *Node) — a self-cycle. Its
// TypeID is nominal (PkgPath+Name), so repeated calls yield distinct
// objects sharing one TypeID.
func buildSelfReferentialType(pkgPath string) *DeclaredType {
	nodeDT := &DeclaredType{PkgPath: pkgPath, Name: "Node"}
	nodeDT.Base = &StructType{
		PkgPath: pkgPath,
		Fields: []FieldType{
			{Name: "key", Type: StringType},
			{Name: "value", Type: IntType},
			{Name: "leftNode", Type: &PointerType{Elt: nodeDT}},
			{Name: "rightNode", Type: &PointerType{Elt: nodeDT}},
		},
	}
	return nodeDT
}

// BenchmarkAssertTypeIsPublic_ColdAcyclic measures the cost paid the first
// time a given type is checked in a process (cache miss): the full
// type-graph walk. The cache is cleared each iteration so every call is a
// miss.
func BenchmarkAssertTypeIsPublic_ColdAcyclic(b *testing.B) {
	store := NewStore(nil, nil, nil)
	pkgPaths, rootPath := registerPublicPkgs(store, "gno.land/r/bench_cold", 5)
	rlm := NewRealm(rootPath)
	root := buildPublicType(pkgPaths, rootPath, 20)
	cache := storeTypePrivacyCache(store)

	for b.Loop() {
		cache.m = make(map[TypeID]bool) // force a cold cache each iteration
		rlm.assertTypeIsPublic(store, root, map[TypeID]struct{}{})
	}
}

// BenchmarkAssertTypeIsPublic_WarmAcyclic measures the cost paid on every
// commit after the first: a TypeID already in the store cache — a hit.
// This is the case the redesign makes fast across transactions; the
// previous design could not, because its memo lived on the per-tx object
// and never survived the commit.
func BenchmarkAssertTypeIsPublic_WarmAcyclic(b *testing.B) {
	store := NewStore(nil, nil, nil)
	pkgPaths, rootPath := registerPublicPkgs(store, "gno.land/r/bench_warm", 5)
	rlm := NewRealm(rootPath)
	root := buildPublicType(pkgPaths, rootPath, 20)
	rlm.assertTypeIsPublic(store, root, map[TypeID]struct{}{}) // warm once

	for b.Loop() {
		rlm.assertTypeIsPublic(store, root, map[TypeID]struct{}{})
	}
}

// BenchmarkAssertTypeIsPublic_ColdSelfReferential is the avl.Node-shaped
// cold counterpart — the first check of a self-referential type.
func BenchmarkAssertTypeIsPublic_ColdSelfReferential(b *testing.B) {
	store := NewStore(nil, nil, nil)
	const pkgPath = "gno.land/p/bench_selfref_cold"
	store.SetCachePackage(&PackageValue{PkgPath: pkgPath, Private: false})
	rlm := NewRealm(pkgPath)
	node := buildSelfReferentialType(pkgPath)
	cache := storeTypePrivacyCache(store)

	for b.Loop() {
		cache.m = make(map[TypeID]bool)
		rlm.assertTypeIsPublic(store, node, map[TypeID]struct{}{})
	}
}

// BenchmarkAssertTypeIsPublic_WarmSelfReferential shows that, unlike the
// previous design (which never cached any type reachable through a cycle,
// so avl.Tree-backed realms got zero benefit), self-referential types now
// hit the cache across commits like any other.
func BenchmarkAssertTypeIsPublic_WarmSelfReferential(b *testing.B) {
	store := NewStore(nil, nil, nil)
	const pkgPath = "gno.land/p/bench_selfref_warm"
	store.SetCachePackage(&PackageValue{PkgPath: pkgPath, Private: false})
	rlm := NewRealm(pkgPath)
	node := buildSelfReferentialType(pkgPath)
	rlm.assertTypeIsPublic(store, node, map[TypeID]struct{}{}) // warm once

	for b.Loop() {
		rlm.assertTypeIsPublic(store, node, map[TypeID]struct{}{})
	}
}
