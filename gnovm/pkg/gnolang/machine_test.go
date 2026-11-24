package gnolang

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
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

func TestRunMemPackageBlankTypeDeclNoSlot(t *testing.T) {
	// A blank type decl must not reserve a block slot: persisted package
	// blocks are keyed by name index, so a `type _ ...` slot would shift
	// every later name in the block. Reserve must ignore the blank
	// identifier just like Define2 does.
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
			{Name: "a.gno", Body: "package std\n\ntype _ struct{}\n\nvar a = 1\n"},
		},
	}, true)
	pn := store.GetBlockNode(PackageNodeLocation("std")).(*PackageNode)
	assert.NotContains(t, pn.GetBlockNames(), Name(blankIdentifier))
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

func TestDeferGas(t *testing.T) {
	for _, tc := range []struct {
		name             string
		params, captures int
	}{
		{"mixed", 2, 4},
		{"capture_only", 0, 4},
	} {
		for _, unwinding := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/panic=%t", tc.name, unwinding), func(t *testing.T) {
				m := NewMachineWithOptions(MachineOptions{PkgPath: "bench", Output: io.Discard})
				defer m.Release()
				outer := &FuncValue{Type: &FuncType{}, IsClosure: true, Source: &FuncDecl{}, PkgPath: "bench"}
				// Keep a caller below the returning frame for panic propagation.
				for range 2 {
					m.PushValue(TypedValue{T: outer.Type, V: outer})
					m.PushFrameCall(&CallExpr{}, outer, TypedValue{}, false)
					m.PushBlock(&Block{})
				}
				params := make([]FieldType, tc.params)
				args := make([]TypedValue, tc.params)
				for i := range params {
					params[i] = FieldType{Type: IntType}
					args[i] = TypedValue{T: IntType}
				}
				fd := &FuncDecl{}
				fd.NumNames = uint16(tc.params + tc.captures)
				fv := &FuncValue{Type: &FuncType{Params: params}, IsClosure: true, Source: fd, PkgPath: "bench"}
				for range tc.captures {
					fv.Captures = append(fv.Captures, TypedValue{T: heapItemType{}, V: m.Alloc.NewHeapItem(nil, TypedValue{T: IntType})})
				}
				m.LastFrame().PushDefer(Defer{Callable: fv, Args: args, Source: &DeferStmt{}})
				if unwinding {
					m.Exception = &Exception{Value: typedString("unwind")}
					m.doOpPanic2()
				} else {
					m.PushOp(OpReturnCallDefers)
				}
				// Enough for dispatch, one gas short of the parameter/capture charge.
				// Leave allocation unmetered to isolate the CPU charge's ordering.
				limit := int64(OpCPUReturnCallDefers+
					OpCPUSlopeCallParam*tc.params+OpCPUSlopeCallCapture*tc.captures)*GasFactorCPU - 1
				m.GasMeter = store.NewGasMeter(limit)
				frames, blocks := len(m.Frames), len(m.Blocks)
				require.PanicsWithValue(t, stypes.OutOfGasError{Descriptor: "CPUCycles"}, func() { m.Run(StageRun) })
				require.Len(t, m.Frames, frames, "charge must precede PushFrameCall")
				require.Len(t, m.Blocks, blocks, "charge must precede capture copying")
			})
		}
	}
}

// TestCallGas verifies doOpCall charges the parameter and capture CPU cost
// before acquiring the callee block, so an out-of-gas stops before that work.
func TestCallGas(t *testing.T) {
	for _, tc := range []struct {
		name             string
		params, captures int
	}{
		{"mixed", 2, 4},
		{"capture_only", 0, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMachineWithOptions(MachineOptions{PkgPath: "bench", Output: io.Discard})
			defer m.Release()

			params := make([]FieldType, tc.params)
			for i := range params {
				params[i] = FieldType{Type: IntType}
			}
			fd := &FuncDecl{}
			fd.NumNames = uint16(tc.params + tc.captures)
			fv := &FuncValue{Type: &FuncType{Params: params}, IsClosure: true, Source: fd, PkgPath: "bench", body: []Stmt{}}
			for range tc.captures {
				fv.Captures = append(fv.Captures, TypedValue{T: heapItemType{}, V: m.Alloc.NewHeapItem(nil, TypedValue{T: IntType})})
			}

			m.PushValue(TypedValue{T: fv.Type, V: fv})
			m.PushFrameCall(&CallExpr{}, fv, TypedValue{}, false)
			m.PushOp(OpCall)

			// One gas short of the parameter and capture charge. The block
			// (OpCPUAcquireBlock) is left unmetered so the panic can only come
			// from the charge under test, not block setup.
			limit := int64(OpCPUCall+
				OpCPUSlopeCallParam*tc.params+OpCPUSlopeCallCapture*tc.captures)*GasFactorCPU - 1
			m.GasMeter = store.NewGasMeter(limit)
			blocks := len(m.Blocks)

			require.PanicsWithValue(t, stypes.OutOfGasError{Descriptor: "CPUCycles"}, func() { m.Run(StageRun) })
			require.Len(t, m.Blocks, blocks, "charge must precede block acquisition and capture copying")
		})
	}
}

// A deferred nil func must still pop its call's arguments. Otherwise each
// `defer f(...)` leaves them on m.Values, uncharged, until the frame returns.
func TestDeferNilFuncPopsArgs(t *testing.T) {
	m := NewMachineWithOptions(MachineOptions{PkgPath: "main", Output: io.Discard})
	defer m.Release()
	m.RunFiles(m.MustParseFile("main.go", `package main

var f func(a, b, c, d int)

func run() {
	defer func() { recover() }()
	for i := 0; i < 1000; i++ {
		defer f(i, i, i, i)
	}
}

func main() { run() }`))
	m.RunMain()
	assert.Less(t, cap(m.Values), 1000, "value stack grew with the nil defers' arguments")
}
