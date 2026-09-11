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

func TestMachineReleaseClearsBlockPool(t *testing.T) {
	// Gas charging in acquireBlock differs between pool hits and misses, so
	// consensus safety requires the pool to start empty on every run: a warm
	// pool carried across Release (machines are reused via machinePool) would
	// make gas depend on prior machine use.
	m := NewMachineWithOptions(MachineOptions{})
	m.blockPool = append(m.blockPool, &Block{Values: make([]TypedValue, 0, blockPoolValueCap)})
	m.Release()
	assert.Empty(t, m.blockPool, "Machine.Release must not preserve blockPool")
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

// A realm record handed to RunMemPackageOverRealm has to be the one persisted
// at the package's own path, and the run has to save it. Neither is checkable
// after the fact: the ObjectIDs are minted off whichever counter arrives, and
// the record is written back under its own path.
func TestRunMemPackageOverRealmRefusesAForeignRecord(t *testing.T) {
	const pkgPath = "gno.land/r/demo/over"
	mpkg := func() *std.MemPackage {
		return &std.MemPackage{
			Type: MPUserAll,
			Name: "over",
			Path: pkgPath,
			Files: []*std.MemFile{
				{Name: "over.gno", Body: "package over\n\nfunc Hi() string { return \"hi\" }\n"},
			},
		}
	}

	cases := map[string]struct {
		save  bool
		prior *Realm
		want  string
	}{
		"unsaved run": {
			save:  false,
			prior: NewRealm(pkgPath),
			want: "prior realm gno.land/r/demo/over requires save: " +
				"an unsaved run must not touch a persisted realm",
		},
		"another path's record": {
			save:  true,
			prior: NewRealm("gno.land/r/demo/elsewhere"),
			want:  "prior realm gno.land/r/demo/elsewhere is not the realm of package gno.land/r/demo/over",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			db := memdb.NewMemDB()
			store := NewStore(nil,
				dbadapter.StoreConstructor(db, stypes.StoreOptions{}),
				iavl.StoreConstructor(db, stypes.StoreOptions{}))
			m := NewMachine("over", store)
			defer m.Release()
			assert.PanicsWithValue(t, tc.want, func() {
				m.RunMemPackageOverRealm(mpkg(), tc.save, tc.prior)
			})
		})
	}
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
    Realm:
      std
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
    Realm:
      std
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
