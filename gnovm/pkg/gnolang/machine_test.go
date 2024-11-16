package gnolang_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
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

func TestMachineTestMemPackage(t *testing.T) {
	matchFunc := func(pat, str string) (bool, error) { return true, nil }

	tests := []struct {
		name          string
		path          string
		shouldSucceed bool
	}{
		{
			name:          "TestSuccess",
			path:          "testdata/TestMemPackage/success",
			shouldSucceed: true,
		},
		{
			name:          "TestFail",
			path:          "testdata/TestMemPackage/fail",
			shouldSucceed: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// NOTE: Because the purpose of this test is to ensure testing.T.Failed()
			// returns true if a gno test is failing, and because we don't want this
			// to affect the current testing.T, we are creating an other one thanks
			// to testing.RunTests() function.
			testing.RunTests(matchFunc, []testing.InternalTest{
				{
					Name: tt.name,
					F: func(t2 *testing.T) { //nolint:thelper
						rootDir := filepath.Join("..", "..", "..")
						_, store := test.Store(rootDir, false, os.Stdin, os.Stdout, os.Stderr)
						store.SetLogStoreOps(true)
						m := gno.NewMachineWithOptions(gno.MachineOptions{
							PkgPath: "test",
							Output:  os.Stdout,
							Store:   store,
							Context: nil,
						})
						memPkg := gno.ReadMemPackage(tt.path, "test")

						_, _ = memPkg, m
						// XXX TODOm.TestMemPackage(t2, memPkg)

						if tt.shouldSucceed {
							assert.False(t, t2.Failed(), "test %q should have succeed", tt.name)
						} else {
							assert.True(t, t2.Failed(), "test %q should have failed", tt.name)
						}
					},
				},
			})
		})
	}
}
