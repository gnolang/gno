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
	t.Parallel()

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
			&std.MemPackage{
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*std.MemFile{
					{
						Name: "hello.gno",
						Body: `
							package hello
							type S struct{}
							func A() S { return S{} }
							func B() S { return A() }`,
					},
				},
			},
			nil,
			nil,
		},
		{
			"WrongReturn",
			&std.MemPackage{
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*std.MemFile{
					{
						Name: "hello.gno",
						Body: `
							package hello
							type S struct{}
							func A() S { return S{} }
							func B() S { return 11 }`,
					},
				},
			},
			nil,
			errContains("cannot use 11"),
		},
		{
			"ParseError",
			&std.MemPackage{
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*std.MemFile{
					{
						Name: "hello.gno",
						Body: `
							package hello!
							func B() int { return 11 }`,
					},
				},
			},
			nil,
			errContains("found '!'"),
		},
		{
			"MultiError",
			&std.MemPackage{
				Name: "main",
				Path: "gno.land/p/demo/main",
				Files: []*std.MemFile{
					{
						Name: "hello.gno",
						Body: `
							package main
							func main() {
								_, _ = 11
								return 88, 88
							}`,
					},
				},
			},
			nil,
			errContains("assignment mismatch", "too many return values"),
		},
		{
			"TestsIgnored",
			&std.MemPackage{
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*std.MemFile{
					{
						Name: "hello.gno",
						Body: `
							package hello
							func B() int { return 11 }`,
					},
					{
						Name: "hello_test.gno",
						Body: `This is not valid Gno code, but it doesn't matter because test
				files are not checked.`,
					},
				},
			},
			nil,
			nil,
		},
		{
			"ImportFailed",
			&std.MemPackage{
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*std.MemFile{
					{
						Name: "hello.gno",
						Body: `
							package hello
							import "std"
							func Hello() std.Address { return "hello" }`,
					},
				},
			},
			mockPackageGetter{},
			errContains("import not found: std"),
		},
		{
			"ImportSucceeded",
			&std.MemPackage{
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*std.MemFile{
					{
						Name: "hello.gno",
						Body: `
							package hello
							import "std"
							func Hello() std.Address { return "hello" }`,
					},
				},
			},
			mockPackageGetter{
				&std.MemPackage{
					Name: "std",
					Path: "std",
					Files: []*std.MemFile{
						{
							Name: "std.gno",
							Body: `
								package std
								type Address string`,
						},
					},
				},
			},
			nil,
		},
		{
			"ImportBadIdent",
			&std.MemPackage{
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*std.MemFile{
					{
						Name: "hello.gno",
						Body: `
							package hello
							import "std"
							func Hello() std.Address { return "hello" }`,
					},
				},
			},
			mockPackageGetter{
				&std.MemPackage{
					Name: "a_completely_different_identifier",
					Path: "std",
					Files: []*std.MemFile{
						{
							Name: "std.gno",
							Body: `
								package a_completely_different_identifier
								type Address string`,
						},
					},
				},
			},
			errContains("undefined: std", "a_completely_different_identifier and not used"),
		},
	}

	cacheMpg := mockPackageGetterCounts{
		mockPackageGetter{
			&std.MemPackage{
				Name: "bye",
				Path: "bye",
				Files: []*std.MemFile{
					{
						Name: "bye.gno",
						Body: `
							package bye
							import "std"
							func Bye() std.Address { return "bye" }`,
					},
				},
			},
			&std.MemPackage{
				Name: "std",
				Path: "std",
				Files: []*std.MemFile{
					{
						Name: "std.gno",
						Body: `
							package std
							type Address string`,
					},
				},
			},
		},
		make(map[string]int),
	}

	tt = append(tt, testCase{
		"ImportWithCache",
		// This test will make use of the importer's internal cache for package `std`.
		&std.MemPackage{
			Name: "hello",
			Path: "gno.land/p/demo/hello",
			Files: []*std.MemFile{
				{
					Name: "hello.gno",
					Body: `
						package hello
						import (
							"std"
							"bye"
						)
						func Hello() std.Address { return bye.Bye() }`,
				},
			},
		},
		cacheMpg,
		func(t *testing.T, err error) {
			t.Helper()
			require.NoError(t, err)
			assert.Equal(t, map[string]int{"std": 1, "bye": 1}, cacheMpg.counts)
		},
	})

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := TypeCheckMemPackage(tc.pkg, tc.getter, false)
			if tc.check == nil {
				assert.NoError(t, err)
			} else {
				tc.check(t, err)
			}
		})
	}
}

func TestTypeCheckMemPackage_fmt(t *testing.T) {
	t.Parallel()

	input := `
	package hello
		func Hello(name string) string   {return "hello"  + name
}



`

	pkg := &std.MemPackage{
		Name: "hello",
		Path: "gno.land/p/demo/hello",
		Files: []*std.MemFile{
			{
				Name: "hello.gno",
				Body: input,
			},
		},
	}

	mpkgGetter := mockPackageGetter{}
	err := TypeCheckMemPackage(pkg, mpkgGetter, false)
	assert.NoError(t, err)
	assert.Equal(t, input, pkg.Files[0].Body) // unchanged

	expected := `package hello

func Hello(name string) string {
	return "hello" + name
}
`

	err = TypeCheckMemPackage(pkg, mpkgGetter, true)
	assert.NoError(t, err)
	assert.NotEqual(t, input, pkg.Files[0].Body)
	assert.Equal(t, expected, pkg.Files[0].Body)
}
