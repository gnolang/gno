package gnolang

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
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
	m.RunStatement(StageRun, S(Call(X("Redecl"), 11)))

	// Check last value, assuming it is the result of Redecl.
	v := m.Values[0]
	assert.NotNil(t, v)
	assert.Equal(t, StringKind, v.T.Kind())
	assert.Equal(t, StringValue("1"), v.V)
}

func TestMachineString(t *testing.T) {
	cases := []struct {
		name string
		in   *Machine
		want string
	}{
		{
			"nil Machine",
			nil,
			"Machine:nil",
		},
		{
			"created with defaults",
			NewMachineWithOptions(MachineOptions{}),
			"Machine:\n    Stage: \n    Op: []\n    Values: (len: 0)\n    Exprs:\n    Stmts:\n    Blocks:\n    Blocks (other):\n    Frames:\n",
		},
		{
			"created with store and defaults",
			func() *Machine {
				db := memdb.NewMemDB()
				baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
				iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
				store := NewStore(nil, baseStore, iavlStore)
				return NewMachine("std", store)
			}(),
			"Machine:\n    Stage: \n    Op: []\n    Values: (len: 0)\n    Exprs:\n    Stmts:\n    Blocks:\n    Blocks (other):\n    Frames:\n",
		},
		{
			"filled in",
			func() *Machine {
				db := memdb.NewMemDB()
				baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
				iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
				store := NewStore(nil, baseStore, iavlStore)
				m := NewMachine("std", store)
				m.PushOp(OpHalt)
				m.PushExpr(&BasicLitExpr{
					Kind:  INT,
					Value: "100",
				})
				m.Blocks = make([]*Block, 1, 1)
				m.PushStmts(S(Call(X("Redecl"), 11)))
				return m
			}(),
			"Machine:\n    Stage: \n    Op: [OpHalt]\n    Values: (len: 0)\n    Exprs:\n          #0 100\n    Stmts:\n          #0 Redecl<VPUverse(0)>(11)\n    Blocks:\n    Blocks (other):\n    Frames:\n",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
