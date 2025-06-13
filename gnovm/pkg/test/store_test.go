package test_test

import (
	// "os"
	"path/filepath"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func TestStoreWithOptions_GetPackageBranches(t *testing.T) {
	// rootDir must point to gnovm/tests/integ/test/store
	rootDir := filepath.Join("..", "..", "tests", "integ", "test", "store")

	// buildStore calls the real StoreWithOptions and returns the
	// configured gno.Store, with its PackageGetter already wired up.
	makeStore := func(opts test.StoreOptions) gno.Store {
		io := commands.NewTestIO()
		_, store := test.StoreWithOptions(rootDir, io.Err(), opts)
		return store
	}

	// Each case drives one branch in getPackage or in the private
	// _processMemPackage (via PreprocessOnly toggles).
	cases := []struct {
		name      string
		opts      test.StoreOptions
		pkgPath   string
		wantPanic string
	}{
		{
			name:      "empty_path_panics",
			opts:      test.StoreOptions{},
			pkgPath:   "",
			wantPanic: "invalid zero package path in testStore().pkgGetter", // if pkgPath=="", getPackage panics
		},
		{
			name:    "extern_pre_true",
			opts:    test.StoreOptions{WithExtern: true, PreprocessOnly: true},
			pkgPath: "github.com/gnolang/gno/_test/foo", // triggers extern/foo → _processMemPackage with PreprocessOnly
		},
		{
			name:    "extern_pre_false",
			opts:    test.StoreOptions{WithExtern: true, PreprocessOnly: false},
			pkgPath: "github.com/gnolang/gno/_test/foo", // extern/foo → _processMemPackage with RunMemPackage
		},
		{
			name:    "example_pre_true",
			opts:    test.StoreOptions{PreprocessOnly: true},
			pkgPath: "pkgbar", // triggers examples/pkgbar → _processMemPackage with PreprocessOnly
		},
		{
			name:    "example_pre_false",
			opts:    test.StoreOptions{PreprocessOnly: false},
			pkgPath: "pkgbar", // examples/pkgbar → _processMemPackage with RunMemPackage
		},
		{
			name:    "stdlib_pre_true",
			opts:    test.StoreOptions{PreprocessOnly: true},
			pkgPath: "mypkg", // loadStdlib’s preprocessOnly branch near end
		},
		{
			name:    "stdlib_pre_false",
			opts:    test.StoreOptions{PreprocessOnly: false},
			pkgPath: "mypkg", // loadStdlib’s RunMemPackageWithOverrides branch
		},
	}

	for _, tc := range cases {
		tc := tc // capture
		t.Run(tc.name, func(t *testing.T) {
			store := makeStore(tc.opts)

			if tc.wantPanic != "" {
				defer func() {
					if r := recover(); r != nil {
						if rStr, ok := r.(string); ok {
							if tc.wantPanic != rStr {
								t.Errorf("%s: expected %#v panic, got %#v", tc.name, tc.wantPanic, rStr)
							}
						} else {
							t.Errorf("%s: expected string panic, got %#v", tc.name, r)
						}
					} else {
						t.Errorf("%s: expected string panic, got none", tc.name)
					}
				}()
				_ = store.GetPackage(tc.pkgPath, true)
				return
			}

			pv := store.GetPackage(tc.pkgPath, true)
			if pv == nil {
				t.Fatalf("%s: expected non-nil PackageValue; got %v", tc.name, pv)
			}
		})
	}
}
