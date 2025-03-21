package gnolang

import (
	"testing"

	"github.com/gnolang/gno/gnovm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

type mockPackageGetter []*gnovm.MemPackage

func (mi mockPackageGetter) GetMemPackage(path string) *gnovm.MemPackage {
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

func (mpg mockPackageGetterCounts) GetMemPackage(path string) *gnovm.MemPackage {
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
		pkg    *gnovm.MemPackage
		getter MemPackageGetter
		check  func(*testing.T, error)
	}
	tt := []testCase{
		{
			"Simple",
			&gnovm.MemPackage{
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*gnovm.MemFile{
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
			&gnovm.MemPackage{
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*gnovm.MemFile{
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
			&gnovm.MemPackage{
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*gnovm.MemFile{
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
			&gnovm.MemPackage{
				Name: "main",
				Path: "gno.land/p/demo/main",
				Files: []*gnovm.MemFile{
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
			&gnovm.MemPackage{
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*gnovm.MemFile{
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
			&gnovm.MemPackage{
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*gnovm.MemFile{
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
			&gnovm.MemPackage{
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*gnovm.MemFile{
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
				&gnovm.MemPackage{
					Name: "std",
					Path: "std",
					Files: []*gnovm.MemFile{
						{
							Name: "gnovm.gno",
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
			&gnovm.MemPackage{
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*gnovm.MemFile{
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
				&gnovm.MemPackage{
					Name: "a_completely_different_identifier",
					Path: "std",
					Files: []*gnovm.MemFile{
						{
							Name: "gnovm.gno",
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
			&gnovm.MemPackage{
				Name: "bye",
				Path: "bye",
				Files: []*gnovm.MemFile{
					{
						Name: "bye.gno",
						Body: `
							package bye
							import "std"
							func Bye() std.Address { return "bye" }`,
					},
				},
			},
			&gnovm.MemPackage{
				Name: "std",
				Path: "std",
				Files: []*gnovm.MemFile{
					{
						Name: "gnovm.gno",
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
		&gnovm.MemPackage{
			Name: "hello",
			Path: "gno.land/p/demo/hello",
			Files: []*gnovm.MemFile{
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

			format := false
			err := TypeCheckMemPackage(tc.pkg, tc.getter, format)
			if tc.check == nil {
				assert.NoError(t, err)
			} else {
				tc.check(t, err)
			}
		})
	}
}

func TestTypeCheckMemPackage_format(t *testing.T) {
	t.Parallel()

	input := `
	package hello
		func Hello(name string) string   {return "hello"  + name
}



`

	pkg := &gnovm.MemPackage{
		Name: "hello",
		Path: "gno.land/p/demo/hello",
		Files: []*gnovm.MemFile{
			{
				Name: "hello.gno",
				Body: input,
			},
		},
	}

	mpkgGetter := mockPackageGetter{}
	format := false
	err := TypeCheckMemPackage(pkg, mpkgGetter, format)
	assert.NoError(t, err)
	assert.Equal(t, input, pkg.Files[0].Body) // unchanged

	expected := `package hello

func Hello(name string) string {
	return "hello" + name
}
`

	format = true
	err = TypeCheckMemPackage(pkg, mpkgGetter, format)
	assert.NoError(t, err)
	assert.NotEqual(t, input, pkg.Files[0].Body)
	assert.Equal(t, expected, pkg.Files[0].Body)
}

func TestTypeCheckMemPackage_RealmImports(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		pkg    *gnovm.MemPackage
		getter MemPackageGetter
		check  func(*testing.T, error)
	}

	tests := []testCase{
		{
			name: "realm package with init-only imports",
			pkg: &gnovm.MemPackage{
				Name: "gns",
				Path: "gno.land/r/demo/gns",
				Files: []*gnovm.MemFile{
					{
						Name: "gns.gno",
						Body: `
							package gns
							import (
								"std"
								"gno.land/r/demo/registry"
							)

							var (
								adminAddr std.Address
							)
							
							func init() {
								registry.Register("gns")
							}
							
							func GetAdmin() string {
								return string(adminAddr)
							}`,
					},
				},
			},
			getter: mockPackageGetter{
				&gnovm.MemPackage{
					Name: "std",
					Path: "std",
					Files: []*gnovm.MemFile{
						{
							Name: "std.gno",
							Body: `
								package std
								type Address string`,
						},
					},
				},
				&gnovm.MemPackage{
					Name: "registry",
					Path: "gno.land/r/demo/registry",
					Files: []*gnovm.MemFile{
						{
							Name: "registry.gno",
							Body: `
								package registry
								func Register(name string) {}`,
						},
					},
				},
			},
			check: func(t *testing.T, err error) {
				t.Helper()
				require.NoError(t, err, "should not report unused imports in realm packages")
			},
		},
		{
			name: "realm package with cross-realm imports",
			pkg: &gnovm.MemPackage{
				Name: "gns",
				Path: "gno.land/r/demo/gns",
				Files: []*gnovm.MemFile{
					{
						Name: "gns.gno",
						Body: `
							package gns
							import (
								"gno.land/r/demo/token"
								"gno.land/r/demo/registry"
							)
							
							func init() {
								registry.Register()
							}
							
							func Transfer(t token.Token) {
								t.Transfer()
							}`,
					},
				},
			},
			getter: mockPackageGetter{
				&gnovm.MemPackage{
					Name: "token",
					Path: "gno.land/r/demo/token",
					Files: []*gnovm.MemFile{
						{
							Name: "token.gno",
							Body: `
								package token
								type Token interface {
									Transfer()
								}`,
						},
					},
				},
				&gnovm.MemPackage{
					Name: "registry",
					Path: "gno.land/r/demo/registry",
					Files: []*gnovm.MemFile{
						{
							Name: "registry.gno",
							Body: `
								package registry
								func Register() {}`,
						},
					},
				},
			},
			check: func(t *testing.T, err error) {
				t.Helper()
				require.NoError(t, err, "should handle cross-realm imports correctly")
			},
		},
		{
			name: "debug realm package import scope",
			pkg: &gnovm.MemPackage{
				Name: "gns",
				Path: "gno.land/r/demo/gns",
				Files: []*gnovm.MemFile{
					{
						Name: "gns.gno",
						Body: `
							package gns
							import "gno.land/r/demo/registry"
							
							func init() {
								registry.Register("test")
							}`,
					},
				},
			},
			getter: &debugPackageGetter{
				mockPackageGetter: mockPackageGetter{
					&gnovm.MemPackage{
						Name: "registry",
						Path: "gno.land/r/demo/registry",
						Files: []*gnovm.MemFile{
							{
								Name: "registry.gno",
								Body: `
									package registry
									func Register(name string) {}`,
							},
						},
					},
				},
			},
			check: func(t *testing.T, err error) {
				t.Helper()
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := TypeCheckMemPackage(tc.pkg, tc.getter, false)
			if tc.check != nil {
				tc.check(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTypeCheckMemPackage_InitAndCallbacks(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		pkg    *gnovm.MemPackage
		getter MemPackageGetter
		check  func(*testing.T, error)
	}

	tests := []testCase{
		{
			name: "callback registration in init",
			pkg: &gnovm.MemPackage{
				Name: "callback",
				Path: "gno.land/r/demo/callback",
				Files: []*gnovm.MemFile{
					{
						Name: "callback.gno",
						Body: `
							package callback
							import "gno.land/r/demo/events"
							
							func MintCallback(amount uint64) {
								// will be called at runtime
							}
							
							func init() {
								events.RegisterCallback("mint", MintCallback)
							}`,
					},
				},
			},
			getter: mockPackageGetter{
				&gnovm.MemPackage{
					Name: "events",
					Path: "gno.land/r/demo/events",
					Files: []*gnovm.MemFile{
						{
							Name: "events.gno",
							Body: `
								package events
								
								type CallbackFn func(uint64)
								
								func RegisterCallback(event string, fn CallbackFn) {}`,
						},
					},
				},
			},
			check: func(t *testing.T, err error) {
				t.Helper()
				require.NoError(t, err, "callback registration should be recognized as package usage")
			},
		},
		{
			name: "multiple aliases with init usage",
			pkg: &gnovm.MemPackage{
				Name: "main",
				Path: "gno.land/r/demo/main",
				Files: []*gnovm.MemFile{
					{
						Name: "main.gno",
						Body: `
							package main
							
							import (
								ev "gno.land/r/demo/events"
								cb "gno.land/r/demo/callback"
							)
							
							func ProcessEvent(amount uint64) {
								// will be called at runtime
							}
							
							func init() {
								ev.RegisterCallback("process", ProcessEvent)
								cb.SetHandler(ProcessEvent)
							}`,
					},
				},
			},
			getter: mockPackageGetter{
				&gnovm.MemPackage{
					Name: "events",
					Path: "gno.land/r/demo/events",
					Files: []*gnovm.MemFile{
						{
							Name: "events.gno",
							Body: `
								package events
								type CallbackFn func(uint64)
								func RegisterCallback(event string, fn CallbackFn) {}`,
						},
					},
				},
				&gnovm.MemPackage{
					Name: "callback",
					Path: "gno.land/r/demo/callback",
					Files: []*gnovm.MemFile{
						{
							Name: "callback.gno",
							Body: `
								package callback
								type HandlerFn func(uint64)
								func SetHandler(fn HandlerFn) {}`,
						},
					},
				},
			},
			check: func(t *testing.T, err error) {
				t.Helper()
				require.NoError(t, err, "aliased imports in init should be recognized")
			},
		},
		{
			name: "init with deferred callback setup",
			pkg: &gnovm.MemPackage{
				Name: "deferred",
				Path: "gno.land/r/demo/deferred",
				Files: []*gnovm.MemFile{
					{
						Name: "deferred.gno",
						Body: `
							package deferred
							
							import handler "gno.land/r/demo/handler"
							
							var setupCallback func()
							
							func init() {
								setupCallback = func() {
									handler.Register("deferred", processEvent)
								}
								setupCallback()
							}
							
							func processEvent() {}`,
					},
				},
			},
			getter: mockPackageGetter{
				&gnovm.MemPackage{
					Name: "handler",
					Path: "gno.land/r/demo/handler",
					Files: []*gnovm.MemFile{
						{
							Name: "handler.gno",
							Body: `
								package handler
								func Register(name string, fn func()) {}`,
						},
					},
				},
			},
			check: func(t *testing.T, err error) {
				t.Helper()
				require.NoError(t, err, "deferred callback setup should be recognized")
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			debugGetter := &debugPackageGetter{
				mockPackageGetter: tc.getter.(mockPackageGetter),
			}

			err := TypeCheckMemPackage(tc.pkg, debugGetter, false)

			t.Logf("Imported packages: %v", debugGetter.GetImportedPackages())

			if tc.check != nil {
				tc.check(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type debugPackageGetter struct {
	mockPackageGetter
	importedPkgs map[string]bool
}

func (d *debugPackageGetter) GetMemPackage(path string) *gnovm.MemPackage {
	if d.importedPkgs == nil {
		d.importedPkgs = make(map[string]bool)
	}
	pkg := d.mockPackageGetter.GetMemPackage(path)
	if pkg != nil {
		d.importedPkgs[path] = true
	}
	return pkg
}

func (d *debugPackageGetter) GetImportedPackages() []string {
	pkgs := make([]string, 0, len(d.importedPkgs))
	for pkg := range d.importedPkgs {
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}
