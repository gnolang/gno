package gnolang

import (
	"bytes"
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

// TestPreprocessAllFilesAndSaveBlockNodes_SkipsNilMemPackage simulates a
// partial AddMemPackage write: the baseStore index entry exists but the
// iavlStore body is missing. Without the defensive nil-skip the Machine
// SIGSEGVs on `nil.Type.(MemPackageType)` inside ParseMemPackage, crash-
// looping the node. With the skip, the loop logs a warning and continues
// so the node boots and an operator can repair the store.
func TestPreprocessAllFilesAndSaveBlockNodes_SkipsNilMemPackage(t *testing.T) {
	d1, d2 := memdb.NewMemDB(), memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(d1, stypes.StoreOptions{})
	iavlStore := dbadapter.StoreConstructor(d2, stypes.StoreOptions{})
	store := NewStore(nil, baseStore, iavlStore)

	// Forge: counter=1, index slot points at a path with no iavlStore body.
	// IterMemPackage hits GetMemPackage("missing/path") → nil and yields
	// nil on the channel — exactly the case PR-B must survive.
	baseStore.Set(nil, []byte(backendPackageIndexCtrKey()), []byte("1"))
	baseStore.Set(nil, []byte(backendPackageIndexKey(1)), []byte("missing/path"))

	var out bytes.Buffer
	m := NewMachine("test", store)
	m.Output = &out

	require.NotPanics(t, func() {
		m.PreprocessAllFilesAndSaveBlockNodes()
	}, "must survive nil MemPackage from a partial AddMemPackage")

	logged := out.String()
	assert.Contains(t, logged, "WARNING:")
	assert.Contains(t, logged, "IterMemPackage returned nil")
}
