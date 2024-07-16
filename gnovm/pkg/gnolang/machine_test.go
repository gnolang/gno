package gnolang

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/assert"
)

func BenchmarkCreateNewMachine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := NewMachineWithOptions(MachineOptions{})
		m.Release()
	}
}

func TestRunMemPackageWithOverrides_revertToOld(t *testing.T) {
	// A test to check revertToOld is correctly putting back an old value,
	// after preprocessing fails.
	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
	store := NewStore(nil, baseStore, iavlStore)
	m := NewMachine("std", store)
	m.RunMemPackageWithOverrides(&std.MemPackage{
		Name: "std",
		Path: "std",
		Files: []*std.MemFile{
			{Name: "a.gno", Body: `package std; func Redecl(x int) string { return "1" }`},
		},
	}, true)
	result := func() (p string) {
		defer func() {
			p = fmt.Sprint(recover())
		}()
		m.RunMemPackageWithOverrides(&std.MemPackage{
			Name: "std",
			Path: "std",
			Files: []*std.MemFile{
				{Name: "b.gno", Body: `package std; func Redecl(x int) string { var y string; _, _ = y; return "2" }`},
			},
		}, true)
		return
	}()
	t.Log("panic trying to redeclare invalid func", result)
	m.RunStatement(S(Call(X("Redecl"), 11)))

	// Check last value, assuming it is the result of Redecl.
	v := m.Values[0]
	assert.NotNil(t, v)
	assert.Equal(t, StringKind, v.T.Kind())
	assert.Equal(t, StringValue("1"), v.V)
}
