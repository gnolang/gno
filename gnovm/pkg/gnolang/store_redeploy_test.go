package gnolang

import (
	"io"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

func TestSetTypeRefreshesRedeployedType(t *testing.T) {
	db := memdb.NewMemDB()
	backend := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	const pkgPath = "gno.land/r/demo/settype"
	const typeID = TypeID(pkgPath + ".T")

	deploy := func(store *defaultStore, declaration string) {
		m := NewMachineWithOptions(MachineOptions{Store: store, Output: io.Discard})
		m.RunMemPackage(&std.MemPackage{
			Type: MPUserProd,
			Name: "settype",
			Path: pkgPath,
			Files: []*std.MemFile{{
				Name: "settype.gno",
				Body: "package settype\n" + declaration,
			}},
		}, true)
	}

	deploy(NewStore(nil, backend, backend), "type T struct { A int }")

	redeployStore := NewStore(nil, backend, backend)
	oldType := redeployStore.GetType(typeID).(*DeclaredType)
	require.Len(t, oldType.Base.(*StructType).Fields, 1)
	deploy(redeployStore, "type T struct { A int; B string }")

	freshType := NewStore(nil, backend, backend).GetType(typeID).(*DeclaredType)
	fields := freshType.Base.(*StructType).Fields
	require.Len(t, fields, 2, "fresh store should load the redeployed type layout")
	require.Equal(t, StringKind, fields[1].Type.Kind())
}
