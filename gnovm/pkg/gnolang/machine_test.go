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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		Type: MPStdlibProd,
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
			Type: MPStdlibProd,
			Name: "std",
			Path: "std",
			Files: []*std.MemFile{
				{Name: "b.gno", Body: `package std; func Redecl(x int) string { var y string; _, _ = y; return "2" }`},
			},
		}, true)
		return
	}()
	t.Log("panic trying to redeclare invalid func", result)
	results := m.Eval(Call(X("Redecl"), 11))

	// Check last value, assuming it is the result of Redecl.
	require.Len(t, results, 1)
	v := results[0]
	assert.NotNil(t, v)
	assert.Equal(t, StringKind, v.T.Kind())
	assert.Equal(t, NewStringValue("1"), v.V)
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
			`Machine:
    Stage: $
    Op: []
    Values: (len: 0)
    Exprs:
    Stmts:
    Blocks:
    Blocks (other):
    Frames:
`,
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
			`Machine:
    Stage: $
    Op: []
    Values: (len: 0)
    Exprs:
    Stmts:
    Blocks:
    Blocks (other):
    Frames:
`,
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
				m.Blocks = make([]*Block, 1)
				m.PushStmts(S(Call(X("Redecl"), 11)))
				return m
			}(),
			`Machine:
    Stage: $
    Op: [OpHalt]
    Values: (len: 0)
    Exprs:
          #0 100
    Stmts:
          #0 Redecl<VPInvalid(0)>(11)
    Blocks:
    Blocks (other):
    Frames:
`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.String()
			tt.want = strings.ReplaceAll(tt.want, "$\n", "\n")
			assert.Equal(t, tt.want, got)
		})
	}
}
