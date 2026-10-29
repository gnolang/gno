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
