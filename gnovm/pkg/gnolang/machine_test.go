package gnolang_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gnovm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/assert"
)

func BenchmarkCreateNewMachine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := gno.NewMachineWithOptions(gno.MachineOptions{})
		m.Release()
	}
}

func TestRunMemPackageWithOverrides_revertToOld(t *testing.T) {
	// A test to check revertToOld is correctly putting back an old value,
	// after preprocessing fails.
	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
	store := gno.NewStore(nil, baseStore, iavlStore)
	m := gno.NewMachine("std", store)
	m.RunMemPackageWithOverrides(&gnovm.MemPackage{
		Name: "std",
		Path: "std",
		Files: []*gnovm.MemFile{
			{Name: "a.gno", Body: `package std; func Redecl(x int) string { return "1" }`},
		},
	}, true)
	result := func() (p string) {
		defer func() {
			p = fmt.Sprint(recover())
		}()
		m.RunMemPackageWithOverrides(&gnovm.MemPackage{
			Name: "std",
			Path: "std",
			Files: []*gnovm.MemFile{
				{Name: "b.gno", Body: `package std; func Redecl(x int) string { var y string; _, _ = y; return "2" }`},
			},
		}, true)
		return
	}()
	t.Log("panic trying to redeclare invalid func", result)
	m.RunStatement(gno.S(gno.Call(gno.X("Redecl"), 11)))

	// Check last value, assuming it is the result of Redecl.
	v := m.Values[0]
	assert.NotNil(t, v)
	assert.Equal(t, gno.StringKind, v.T.Kind())
	assert.Equal(t, gno.StringValue("1"), v.V)
}
