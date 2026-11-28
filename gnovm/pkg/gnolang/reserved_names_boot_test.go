package gnolang

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// The reserved-name rule is a deploy-gate rule, and it has to stay one:
// preprocess re-runs every stored package at node boot
// (PreprocessAllFilesAndSaveBlockNodes), so a rejection added there refuses
// source the chain already accepted — an upgraded validator would fail to
// start on a package it has been serving for months.
//
// These two tests are the halves of that. TestCheckDeclNames pins what the
// gate refuses; this one pins that preprocess still accepts every one of
// those shapes, because a package carrying one may already be on chain.
func TestPreprocessAcceptsStoredReservedNames(t *testing.T) {
	// The two shapes master already refused are absent: no stored package
	// can carry them, so preprocess may keep refusing them.
	for _, tc := range []struct{ name, body string }{
		{"param", "func F(len int) int { return len }"},
		{"result", "func F(f float64) (int float64, frac float64) { return 1, 0.5 }"},
		{"uverse param", "func F(address string) string { return address }"},
		{"receiver", "type T struct{}\nfunc (string T) M() {}"},
		{"range", "func F(xs []int) { for len := range xs { _ = len } }"},
		{"type switch", "func F(a any) { switch len := a.(type) { default: _ = len } }"},
		{"func type param", "type S struct{ F func(len int) int }"},
		{"interface method", "type I interface{ M(len int) int }"},
		{"closure param", "func F() { g := func(panic string) {}; _ = g }"},
		{"cur local", "func F() { cur := 1; _ = cur }"},
		{"cur non-realm param", "func F(cur int) int { return cur }"},
		{"cur func", "func cur() {}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := memdb.NewMemDB()
			store := NewStore(nil,
				dbadapter.StoreConstructor(db, stypes.StoreOptions{}),
				iavl.StoreConstructor(db, stypes.StoreOptions{}))
			m := NewMachineWithOptions(MachineOptions{Store: store})
			defer m.Release()
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("preprocess refused a shape the chain may already store: %v", r)
				}
			}()
			m.RunMemPackage(&std.MemPackage{
				Type: MPUserAll, Name: "stored", Path: "gno.land/p/demo/stored",
				Files: []*std.MemFile{{Name: "a.gno", Body: "package stored\n\n" + tc.body + "\n"}},
			}, true)
		})
	}
}

// End to end over the path that actually breaks: a package already in the
// store, re-preprocessed by a restarting node.
func TestBootPreprocessesStoredReservedNames(t *testing.T) {
	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
	store := NewStore(nil, baseStore, iavlStore)

	// Written by a binary that predates the rule.
	store.AddMemPackage(&std.MemPackage{
		Type: MPUserAll, Name: "stored", Path: "gno.land/p/demo/stored",
		Files: []*std.MemFile{{Name: "a.gno", Body: `package stored

func Len(len int) int { return len }

func Cur() int { cur := 1; return cur }
`}},
	}, MPUserAll)

	m := NewMachineWithOptions(MachineOptions{Store: store})
	defer m.Release()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("a restarting node failed to boot on a stored package: %v", r)
		}
	}()
	m.PreprocessAllFilesAndSaveBlockNodes()
}

// The same rule for the `cur` rebind, which is a separate shape: `cur, x :=`
// ASSIGNS the crossing parameter rather than declaring it, so the gate catches
// it through go/types Uses, not Defs. Master accepted it, so a stored package
// may carry it and preprocess must still run it. Found by dora against an
// earlier cut of this change, which refused it in preprocess and so would have
// stopped every upgraded node from booting.
func TestBootStoredRealmTypedCurRebind(t *testing.T) {
	db := memdb.NewMemDB()
	store := NewStore(nil,
		dbadapter.StoreConstructor(db, stypes.StoreOptions{}),
		iavl.StoreConstructor(db, stypes.StoreOptions{}))
	store.AddMemPackage(&std.MemPackage{
		Type: MPUserAll, Name: "stored", Path: "gno.land/r/demo/stored",
		Files: []*std.MemFile{{Name: "a.gno", Body: `package stored

func F(cur realm) string {
	cur, y := cur, 1
	_ = y
	return cur.PkgPath()
}
`}},
	}, MPUserAll)

	m := NewMachineWithOptions(MachineOptions{Store: store})
	defer m.Release()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("a restarting node failed to boot on a stored package: %v", r)
		}
	}()
	m.PreprocessAllFilesAndSaveBlockNodes()
}

// The third shape, and the one that shows why the rule may not sit in
// preprocess even "belt and braces": a crossing function may bind its first
// realm parameter as `_`, and master accepts a local `cur` threaded from it by
// a non-crossing call. Tightening the cur-call provenance check by NAME
// refused that at boot. Found by dora on 348078dbe.
func TestBootStoredCurThreadedFromBlankParam(t *testing.T) {
	const head = "package stored\n\nfunc G(cur realm) string { return cur.PkgPath() }\n\n"
	for _, tc := range []struct{ name, body string }{
		{"func decl", "func F(_ realm, c realm) string {\n\tcur := c\n\treturn G(cur)\n}"},
		{"func literal", "var F = func(_ realm, c realm) string {\n\tcur := c\n\treturn G(cur)\n}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := memdb.NewMemDB()
			store := NewStore(nil,
				dbadapter.StoreConstructor(db, stypes.StoreOptions{}),
				iavl.StoreConstructor(db, stypes.StoreOptions{}))
			store.AddMemPackage(&std.MemPackage{
				Type: MPUserAll, Name: "stored", Path: "gno.land/r/demo/stored",
				Files: []*std.MemFile{{Name: "a.gno", Body: head + tc.body + "\n"}},
			}, MPUserAll)
			m := NewMachineWithOptions(MachineOptions{Store: store})
			defer m.Release()
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("a restarting node failed to boot on a stored package: %v", r)
				}
			}()
			m.PreprocessAllFilesAndSaveBlockNodes()
		})
	}
}
