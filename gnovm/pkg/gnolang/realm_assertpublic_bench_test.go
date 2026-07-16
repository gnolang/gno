package gnolang

import (
	"strconv"
	"testing"
)

// This file is deliberately self-contained (no shared helpers from
// realm_privatedep_test.go) so it can be dropped, unmodified, onto a
// pre-typeHasPrivateDep commit to get a before/after comparison via
// benchstat: both `assertTypeIsPublic` and the Store/PackageValue/
// StructType APIs it uses existed unchanged before that change.

// buildPublicTypeGraph builds a StructType with numFields fields, each of
// a distinct nested StructType (also with a handful of scalar fields),
// spread across numPkgs distinct public packages — representative of a
// mid-sized realm struct that embeds a few imported types. store is
// populated with one public PackageValue per package path used.
func buildPublicTypeGraph(store Store, pkgPrefix string, numPkgs, numFields int) *StructType {
	pkgPaths := make([]string, numPkgs)
	for i := range pkgPaths {
		pkgPaths[i] = pkgPrefix + "/pkg" + strconv.Itoa(i)
		store.SetCachePackage(&PackageValue{PkgPath: pkgPaths[i], Private: false})
	}

	rootPath := pkgPrefix + "/root"
	store.SetCachePackage(&PackageValue{PkgPath: rootPath, Private: false})

	fields := make([]FieldType, numFields)
	for i := range fields {
		nestedPath := pkgPaths[i%numPkgs]
		nested := &StructType{
			PkgPath: nestedPath,
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

// BenchmarkAssertTypeIsPublic_RepeatedCommits simulates saveUnsavedObjects
// calling assertTypeIsPublic with a FRESH visited map on every commit (as
// happens once per real transaction — see saveUnsavedObjects's tids map),
// for the SAME type across many separate commits. This is the exact
// per-transaction re-walk pattern typeHasPrivateDep's cache targets (see
// gnovm/adr/prxxxx_type_privacy_dependency_cache.md): without it, this is
// O(graph size) on every single call; with the cache, every call after
// the first is O(1).
func BenchmarkAssertTypeIsPublic_RepeatedCommits(b *testing.B) {
	store := NewStore(nil, nil, nil)
	root := buildPublicTypeGraph(store, "gno.land/r/bench_repeated", 5, 20)
	rlm := NewRealm(root.PkgPath)

	for b.Loop() {
		visited := map[TypeID]struct{}{}
		rlm.assertTypeIsPublic(store, root, visited)
	}
}

// BenchmarkAssertTypeIsPublic_AlwaysNewType is the adversarial case for
// the cache: every iteration builds and checks a BRAND NEW type graph, so
// typeHasPrivateDep's cache can never hit. This measures the fix's
// worst-case overhead (the walker's extra bookkeeping) rather than its
// benefit, to keep the comparison honest.
func BenchmarkAssertTypeIsPublic_AlwaysNewType(b *testing.B) {
	store := NewStore(nil, nil, nil)
	rlm := NewRealm("gno.land/r/bench_new")

	i := 0
	for b.Loop() {
		root := buildPublicTypeGraph(store, "gno.land/r/bench_new_"+strconv.Itoa(i), 5, 20)
		visited := map[TypeID]struct{}{}
		rlm.assertTypeIsPublic(store, root, visited)
		i++
	}
}

// buildSelfReferentialTypeGraph builds a DeclaredType shaped exactly like
// gno.land/p/nt/avl's Node struct: two fields pointing back at the type
// itself (leftNode/rightNode *Node). This is the shape
// typeHasPrivateDep's cycle-safety rule (see
// gnovm/adr/prxxxx_type_privacy_dependency_cache.md, "Limitations")
// deliberately excludes from the permanent cache.
func buildSelfReferentialTypeGraph(store Store, pkgPath string) *DeclaredType {
	store.SetCachePackage(&PackageValue{PkgPath: pkgPath, Private: false})

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

// BenchmarkAssertTypeIsPublic_RepeatedCommits_SelfReferential is the
// avl.Node-shaped counterpart to BenchmarkAssertTypeIsPublic_
// RepeatedCommits: same type, many separate commits, simulating repeated
// saves of an avl.Tree-backed realm. Unlike that benchmark, this type is
// self-referential, so it's expected to show NO improvement call-to-call
// — every call re-walks the graph from scratch, same as before this
// change. This exists to make that limitation measurable, not just
// documented in the ADR.
func BenchmarkAssertTypeIsPublic_RepeatedCommits_SelfReferential(b *testing.B) {
	store := NewStore(nil, nil, nil)
	node := buildSelfReferentialTypeGraph(store, "gno.land/p/bench_selfref")
	rlm := NewRealm(node.PkgPath)

	for b.Loop() {
		visited := map[TypeID]struct{}{}
		rlm.assertTypeIsPublic(store, node, visited)
	}
}
