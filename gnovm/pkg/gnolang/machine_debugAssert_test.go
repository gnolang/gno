//go:build debugAssert

package gnolang

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// The predefine-time bookkeeping audit must fire when a function-local type
// is missing from ATTR_FUNC_LOCAL_TYPES — i.e. a mint path forgot
// AddFuncLocalType. Guards the completeness invariant the collection
// design rests on.
func TestAssertFuncLocalTypesCompleteFires(t *testing.T) {
	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
	store := NewStore(nil, baseStore, iavlStore)
	m := NewMachine("std", store)
	defer m.Release()
	m.RunMemPackage(&std.MemPackage{
		Type: MPStdlibProd,
		Name: "std",
		Path: "std",
		Files: []*std.MemFile{
			{Name: "a.gno", Body: `package std

func mk() any {
	type S struct{}
	return S{}
}

var X any = mk()
`},
		},
	}, true)
	pv := m.Package
	pn := pv.GetPackageNode(store)
	// Simulate a mint path that forgot AddFuncLocalType.
	pn.DelAttribute(ATTR_FUNC_LOCAL_TYPES)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected the audit to panic, got none")
		}
		if !strings.Contains(fmt.Sprint(r), "not collected at predefine") {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	m.saveFuncLocalTypes(pv)
}
