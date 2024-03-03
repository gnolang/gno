package gnolang

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

func TestParseForLoop(t *testing.T) {
	t.Parallel()

	gocode := `package main
func main(){
	for i:=0; i<10; i++ {
		if i == -1 {
			return
		}
	}
}`
	n, err := ParseFile("main.go", gocode)
	assert.NoError(t, err, "ParseFile error")
	assert.NotNil(t, n, "ParseFile error")
	fmt.Printf("CODE:\n%s\n\n", gocode)
	fmt.Printf("AST:\n%#v\n\n", n)
	fmt.Printf("AST.String():\n%s\n", n.String())
}

func newMemPackage(
	pkgName, pkgPath string,
	namesAndFiles ...string,
) *std.MemPackage {
	if len(namesAndFiles)%2 != 0 {
		panic("namesAndFiles must be pairs")
	}
	files := make([]*std.MemFile, 0, len(namesAndFiles)/2)
	for i := 0; i < len(namesAndFiles); i += 2 {
		files = append(files, &std.MemFile{
			Name: namesAndFiles[i],
			Body: namesAndFiles[i+1],
		})
	}
	return &std.MemPackage{
		Name:  pkgName,
		Path:  pkgPath,
		Files: files,
	}
}

type mockPackageGetter []*std.MemPackage

func (mi mockPackageGetter) GetMemPackage(path string) *std.MemPackage {
	for _, pkg := range mi {
		if pkg.Path == path {
			return pkg
		}
	}
	return nil
}

type mockPackageGetterCounts struct {
	mockPackageGetter
	counts map[string]int
}

func (mpg mockPackageGetterCounts) GetMemPackage(path string) *std.MemPackage {
	mpg.counts[path]++
	return mpg.mockPackageGetter.GetMemPackage(path)
}

func TestTypeCheckMemPackage(t *testing.T) {
	// if len(ss) > 0, then multierr.Errors must decompose it in errors, and
	// each error in order must contain the associated string.
	errContains := func(s0 string, ss ...string) func(*testing.T, error) {
		return func(t *testing.T, err error) {
			t.Helper()
			errs := multierr.Errors(err)
			if len(errs) == 0 {
				t.Errorf("expected an error, got nil")
				return
			}
			want := len(ss) + 1
			if len(errs) != want {
				t.Errorf("expected %d errors, got %d", want, len(errs))
				return
			}
			assert.ErrorContains(t, errs[0], s0)
			for idx, err := range errs[1:] {
				assert.ErrorContains(t, err, ss[idx])
			}
		}
	}

	type testCase struct {
		name   string
		pkg    *std.MemPackage
		getter MemPackageGetter
		check  func(*testing.T, error)
	}
	tt := []testCase{
		{
			"Simple",
			newMemPackage(
				"hello", "gno.land/p/demo/hello",

				"hello.gno",
				`package hello
				type S struct{}
				func A() S { return S{} }
				func B() S { return A() }`,
			),
			nil,
			nil,
		},
		{
			"WrongReturn",
			newMemPackage(
				"hello", "gno.land/p/demo/hello",

				"hello.gno",
				`package hello
				type S struct{}
				func A() S { return S{} }
				func B() S { return 11 }`,
			),
			nil,
			errContains("cannot use 11"),
		},
		{
			"ParseError",
			newMemPackage(
				"hello", "gno.land/p/demo/hello",

				"hello.gno",
				`package hello!
				func B() int { return 11 }`,
			),
			nil,
			errContains("found '!'"),
		},
		{
			"MultiError",
			newMemPackage(
				"main", "gno.land/p/demo/main",

				"hello.gno",
				`package main
					func main() {
						_, _ = 11
						return 88, 88
					}`,
			),
			nil,
			errContains("assignment mismatch", "too many return values"),
		},
		{
			"TestsIgnored",
			newMemPackage(
				"hello", "gno.land/p/demo/hello",

				"hello.gno",
				`package hello
				func B() int { return 11 }`,
				"hello_test.gno",
				`This is not valid Gno code, but it doesn't matter because test
				files are not checked.`,
			),
			nil,
			nil,
		},
		{
			"ImportFailed",
			newMemPackage(
				"hello", "gno.land/p/demo/hello",

				"hello.gno",
				`package hello
				import "std"
				func Hello() std.Address { return "hello" }`,
			),
			mockPackageGetter{},
			errContains("import not found: std"),
		},
		{
			"ImportSucceeded",
			newMemPackage(
				"hello", "gno.land/p/demo/hello",

				"hello.gno",
				`package hello
				import "std"
				func Hello() std.Address { return "hello" }`,
			),
			mockPackageGetter{
				newMemPackage(
					"std", "std",

					"std.gno",
					`package std
					type Address string`,
				),
			},
			nil,
		},
		{
			"ImportBadIdent",
			newMemPackage(
				"hello", "gno.land/p/demo/hello",

				"hello.gno",
				`package hello
				import "std"
				func Hello() std.Address { return "hello" }`,
			),
			mockPackageGetter{
				newMemPackage(
					"a_completely_dfferent_identifier", "std",

					"std.gno",
					`package a_completely_different_identifier
					type Address string`,
				),
			},
			errContains("undefined: std", "a_completely_different_identifier and not used"),
		},
	}

	cacheMpg := mockPackageGetterCounts{
		mockPackageGetter{
			newMemPackage(
				"bye", "bye",

				"bye.gno",
				`package bye
				import "std"
				func Bye() std.Address { return "bye" }`,
			),
			newMemPackage(
				"std", "std",

				"std.gno",
				`package std
				type Address string`,
			),
		},
		make(map[string]int),
	}

	tt = append(tt, testCase{
		"ImportWithCache",
		// This test will make use of the importer's internal cache for package `std`.
		newMemPackage(
			"hello", "gno.land/p/demo/hello",

			"hello.gno",
			`package hello
			import (
				"std"
				"bye"
			)
			func Hello() std.Address { return bye.Bye() }`,
		),
		cacheMpg,
		func(t *testing.T, err error) {
			t.Helper()
			require.NoError(t, err)
			assert.Equal(t, map[string]int{"std": 1, "bye": 1}, cacheMpg.counts)
		},
	})

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := TypeCheckMemPackage(tc.pkg, tc.getter)
			if tc.check == nil {
				assert.NoError(t, err)
			} else {
				tc.check(t, err)
			}
		})
	}
}
