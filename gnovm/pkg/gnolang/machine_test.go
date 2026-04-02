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
	assert.Equal(t, StringValue("1"), v.V)
}

func TestPreprocessMemPackage_recovery(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
	store := NewStore(nil, baseStore, iavlStore)

	t.Run("syntax error is recovered", func(t *testing.T) {
		t.Parallel()

		m := NewMachineWithOptions(MachineOptions{Store: store})
		defer m.Release()

		mpkg := &std.MemPackage{
			Type: MPStdlibProd,
			Name: "broken",
			Path: "broken",
			Files: []*std.MemFile{
				{Name: "broken.gno", Body: `package broken; func }{`},
			},
		}
		err := m.preprocessMemPackage(mpkg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "preprocess broken")
	})

	t.Run("package name mismatch is recovered", func(t *testing.T) {
		t.Parallel()

		m := NewMachineWithOptions(MachineOptions{Store: store})
		defer m.Release()

		mpkg := &std.MemPackage{
			Type: MPStdlibProd,
			Name: "mypkg",
			Path: "mypkg",
			Files: []*std.MemFile{
				{Name: "f.gno", Body: `package wrongname; func X() {}`},
			},
		}
		err := m.preprocessMemPackage(mpkg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "preprocess mypkg")
	})

	t.Run("valid package succeeds", func(t *testing.T) {
		t.Parallel()

		m := NewMachineWithOptions(MachineOptions{Store: store})
		defer m.Release()

		mpkg := &std.MemPackage{
			Type: MPStdlibProd,
			Name: "hello",
			Path: "hello",
			Files: []*std.MemFile{
				{Name: "hello.gno", Body: `package hello; func Hi() string { return "hi" }`},
			},
		}
		err := m.preprocessMemPackage(mpkg)
		assert.NoError(t, err)
	})
}

func TestPreprocessAllFilesAndSaveBlockNodes_skipsBroken(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
	store := NewStore(nil, baseStore, iavlStore)

	// Add a valid package directly to the store.
	store.AddMemPackage(&std.MemPackage{
		Type: MPStdlibAll,
		Name: "good",
		Path: "good",
		Files: []*std.MemFile{
			{Name: "good.gno", Body: `package good; func Hello() string { return "hello" }`},
		},
	}, MPAnyAll)

	// Add a broken package (syntax error) directly to the store.
	// This simulates a package that was valid when deployed but
	// fails preprocessing on restart (e.g. due to a VM update).
	store.AddMemPackage(&std.MemPackage{
		Type: MPStdlibAll,
		Name: "broken",
		Path: "broken",
		Files: []*std.MemFile{
			{Name: "broken.gno", Body: `package broken; func }{`},
		},
	}, MPAnyAll)

	m := NewMachineWithOptions(MachineOptions{Store: store})
	defer m.Release()

	failed := m.PreprocessAllFilesAndSaveBlockNodes()

	// The broken package should be in the failed list.
	assert.Contains(t, failed, "broken")
	// The good package should NOT be in the failed list.
	assert.NotContains(t, failed, "good")
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
