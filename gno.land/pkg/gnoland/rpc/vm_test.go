package rpc

import (
	"errors"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/mock"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_VMEval(t *testing.T) {
	t.Parallel()

	t.Run("invalid context creation", func(t *testing.T) {
		t.Parallel()

		var (
			sdkErr = errors.New("context err")
			app    = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					return sdk.Context{}, sdkErr
				},
			}

			server = NewServer(app, log.NewNoopLogger())

			height = int64(0)
			realm  = "gno.land/r/example"
			expr   = "Func()"
		)

		result, err := server.VMEval(nil, height, realm, expr)
		require.Empty(t, result)

		assert.ErrorIs(t, err, sdkErr)
	})

	t.Run("invalid eval", func(t *testing.T) {
		t.Parallel()

		var (
			queryErr = errors.New("query err")

			expectedRealm  = "gno.land/r/example"
			expectedExpr   = "Func()"
			expectedHeight = int64(10)

			keeper = &mock.VMKeeper{
				QueryEvalFn: func(_ sdk.Context, realm string, expr string) (string, error) {
					require.Equal(t, expectedRealm, realm)
					require.Equal(t, expectedExpr, expr)

					return "", queryErr
				},
			}
			app = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					require.Equal(t, expectedHeight, height)

					return sdk.Context{}, nil
				},
				vmKeeperFn: func() vm.VMKeeperI {
					return keeper
				},
			}

			server = NewServer(app, log.NewNoopLogger())
		)

		result, err := server.VMEval(nil, expectedHeight, expectedRealm, expectedExpr)
		require.Empty(t, result)

		assert.ErrorIs(t, err, queryErr)
	})

	t.Run("valid eval", func(t *testing.T) {
		t.Parallel()

		var (
			expectedRealm  = "gno.land/r/example"
			expectedExpr   = "Func()"
			expectedResult = "hello"
			expectedHeight = int64(10)

			keeper = &mock.VMKeeper{
				QueryEvalFn: func(_ sdk.Context, realm string, expr string) (string, error) {
					require.Equal(t, expectedRealm, realm)
					require.Equal(t, expectedExpr, expr)

					return expectedResult, nil
				},
			}
			app = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					require.Equal(t, expectedHeight, height)

					return sdk.Context{}, nil
				},
				vmKeeperFn: func() vm.VMKeeperI {
					return keeper
				},
			}

			server = NewServer(app, log.NewNoopLogger())
		)

		result, err := server.VMEval(nil, expectedHeight, expectedRealm, expectedExpr)
		require.NoError(t, err)

		assert.Equal(t, expectedResult, result)
	})
}

func TestServer_VMRender(t *testing.T) {
	t.Parallel()

	t.Run("invalid context creation", func(t *testing.T) {
		t.Parallel()

		var (
			sdkErr = errors.New("context err")
			app    = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					return sdk.Context{}, sdkErr
				},
			}

			server = NewServer(app, log.NewNoopLogger())

			height = int64(0)
			realm  = "gno.land/r/example"
		)

		result, err := server.VMRender(nil, height, realm, "")
		require.Empty(t, result)

		assert.ErrorIs(t, err, sdkErr)
	})

	t.Run("invalid render", func(t *testing.T) {
		t.Parallel()

		var (
			queryErr = errors.New("query err")

			expectedRealm  = "gno.land/r/example"
			expectedExpr   = "Render(\"\")"
			expectedHeight = int64(10)

			keeper = &mock.VMKeeper{
				QueryEvalFn: func(_ sdk.Context, realm string, expr string) (string, error) {
					require.Equal(t, expectedRealm, realm)
					require.Equal(t, expectedExpr, expr)

					return "", queryErr
				},
			}
			app = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					require.Equal(t, expectedHeight, height)

					return sdk.Context{}, nil
				},
				vmKeeperFn: func() vm.VMKeeperI {
					return keeper
				},
			}

			server = NewServer(app, log.NewNoopLogger())
		)

		result, err := server.VMRender(nil, expectedHeight, expectedRealm, "")
		require.Empty(t, result)

		assert.ErrorIs(t, err, queryErr)
	})

	t.Run("valid render", func(t *testing.T) {
		t.Parallel()

		var (
			expectedRealm  = "gno.land/r/example"
			expectedExpr   = "Render(\"\")"
			expectedResult = "hello render"
			expectedHeight = int64(10)

			keeper = &mock.VMKeeper{
				QueryEvalFn: func(_ sdk.Context, realm string, expr string) (string, error) {
					require.Equal(t, expectedRealm, realm)
					require.Equal(t, expectedExpr, expr)

					return expectedResult, nil
				},
			}
			app = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					require.Equal(t, expectedHeight, height)

					return sdk.Context{}, nil
				},
				vmKeeperFn: func() vm.VMKeeperI {
					return keeper
				},
			}

			server = NewServer(app, log.NewNoopLogger())
		)

		result, err := server.VMRender(nil, expectedHeight, expectedRealm, "")
		require.NoError(t, err)

		assert.Equal(t, expectedResult, result)
	})
}

func TestServer_VMFuncs(t *testing.T) {
	t.Parallel()

	t.Run("invalid context creation", func(t *testing.T) {
		t.Parallel()

		var (
			sdkErr = errors.New("context err")
			app    = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					return sdk.Context{}, sdkErr
				},
			}

			server = NewServer(app, log.NewNoopLogger())

			height = int64(0)
			realm  = "gno.land/r/example"
		)

		result, err := server.VMFuncs(nil, height, realm)
		require.Empty(t, result)

		assert.ErrorIs(t, err, sdkErr)
	})

	t.Run("invalid funcs", func(t *testing.T) {
		t.Parallel()

		var (
			queryErr = errors.New("query err")

			expectedRealm  = "gno.land/r/example"
			expectedHeight = int64(10)

			keeper = &mock.VMKeeper{
				QueryFuncsFn: func(_ sdk.Context, realm string) (vm.FunctionSignatures, error) {
					require.Equal(t, expectedRealm, realm)

					return nil, queryErr
				},
			}
			app = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					require.Equal(t, expectedHeight, height)

					return sdk.Context{}, nil
				},
				vmKeeperFn: func() vm.VMKeeperI {
					return keeper
				},
			}

			server = NewServer(app, log.NewNoopLogger())
		)

		result, err := server.VMFuncs(nil, expectedHeight, expectedRealm)
		require.Empty(t, result)

		assert.ErrorIs(t, err, queryErr)
	})

	t.Run("valid funcs", func(t *testing.T) {
		t.Parallel()

		var (
			expectedRealm    = "gno.land/r/example"
			expectedFuncSigs = vm.FunctionSignatures{
				{
					FuncName: "hello1",
				},
				{
					FuncName: "hello2",
				},
			}
			expectedHeight = int64(10)

			keeper = &mock.VMKeeper{
				QueryFuncsFn: func(_ sdk.Context, realm string) (vm.FunctionSignatures, error) {
					require.Equal(t, expectedRealm, realm)

					return expectedFuncSigs, nil
				},
			}
			app = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					require.Equal(t, expectedHeight, height)

					return sdk.Context{}, nil
				},
				vmKeeperFn: func() vm.VMKeeperI {
					return keeper
				},
			}

			server = NewServer(app, log.NewNoopLogger())
		)

		result, err := server.VMFuncs(nil, expectedHeight, expectedRealm)
		require.NoError(t, err)

		assert.Equal(t, expectedFuncSigs.JSON(), result)
	})
}

func TestServer_VMPaths(t *testing.T) {
	t.Parallel()

	t.Run("invalid context creation", func(t *testing.T) {
		t.Parallel()

		var (
			sdkErr = errors.New("context err")
			app    = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					return sdk.Context{}, sdkErr
				},
			}

			server = NewServer(app, log.NewNoopLogger())

			height = int64(0)
			realm  = "gno.land/r/example"
		)

		result, err := server.VMPaths(nil, height, realm, 1)
		require.Empty(t, result)

		assert.ErrorIs(t, err, sdkErr)
	})

	t.Run("invalid paths", func(t *testing.T) {
		t.Parallel()

		var (
			queryErr = errors.New("query err")

			expectedTarget = "gno.land/r/example"
			expectedLimit  = 10
			expectedHeight = int64(10)

			keeper = &mock.VMKeeper{
				QueryPathsFn: func(_ sdk.Context, target string, limit int) ([]string, error) {
					require.Equal(t, expectedTarget, target)
					require.Equal(t, expectedLimit, limit)

					return nil, queryErr
				},
			}
			app = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					require.Equal(t, expectedHeight, height)

					return sdk.Context{}, nil
				},
				vmKeeperFn: func() vm.VMKeeperI {
					return keeper
				},
			}

			server = NewServer(app, log.NewNoopLogger())
		)

		result, err := server.VMPaths(nil, expectedHeight, expectedTarget, expectedLimit)
		require.Empty(t, result)

		assert.ErrorIs(t, err, queryErr)
	})

	t.Run("valid paths", func(t *testing.T) {
		t.Parallel()

		var (
			expectedTarget = "gno.land/r/example"
			expectedLimit  = 10
			expectedPaths  = []string{expectedTarget}
			expectedHeight = int64(10)

			keeper = &mock.VMKeeper{
				QueryPathsFn: func(_ sdk.Context, target string, limit int) ([]string, error) {
					require.Equal(t, expectedTarget, target)
					require.Equal(t, expectedLimit, limit)

					return expectedPaths, nil
				},
			}
			app = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					require.Equal(t, expectedHeight, height)

					return sdk.Context{}, nil
				},
				vmKeeperFn: func() vm.VMKeeperI {
					return keeper
				},
			}

			server = NewServer(app, log.NewNoopLogger())
		)

		result, err := server.VMPaths(nil, expectedHeight, expectedTarget, expectedLimit)
		require.NoError(t, err)

		assert.Equal(t, strings.Join(expectedPaths, "\n"), result)
	})
}

func TestServer_VMFile(t *testing.T) {
	t.Parallel()

	t.Run("invalid context creation", func(t *testing.T) {
		t.Parallel()

		var (
			sdkErr = errors.New("context err")
			app    = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					return sdk.Context{}, sdkErr
				},
			}

			server = NewServer(app, log.NewNoopLogger())

			height   = int64(0)
			filepath = "gno.land/r/example/file.gno"
		)

		result, err := server.VMFile(nil, height, filepath)
		require.Empty(t, result)

		assert.ErrorIs(t, err, sdkErr)
	})

	t.Run("invalid file query", func(t *testing.T) {
		t.Parallel()

		var (
			queryErr       = errors.New("query err")
			expectedPath   = "gno.land/r/example/file.gno"
			expectedHeight = int64(0)

			keeper = &mock.VMKeeper{
				QueryFileFn: func(_ sdk.Context, fp string) (string, error) {
					require.Equal(t, expectedPath, fp)

					return "", queryErr
				},
			}
			app = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					require.Equal(t, expectedHeight, height)

					return sdk.Context{}, nil
				},
				vmKeeperFn: func() vm.VMKeeperI {
					return keeper
				},
			}

			server = NewServer(app, log.NewNoopLogger())
		)

		result, err := server.VMFile(nil, expectedHeight, expectedPath)
		require.Empty(t, result)

		assert.ErrorIs(t, err, queryErr)
	})

	t.Run("valid file query", func(t *testing.T) {
		t.Parallel()

		var (
			expectedPath   = "gno.land/r/example/file.gno"
			expectedHeight = int64(0)
			expectedResult = "file contents"

			keeper = &mock.VMKeeper{
				QueryFileFn: func(_ sdk.Context, path string) (string, error) {
					require.Equal(t, expectedPath, path)

					return expectedResult, nil
				},
			}
			app = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					require.Equal(t, expectedHeight, height)

					return sdk.Context{}, nil
				},
				vmKeeperFn: func() vm.VMKeeperI {
					return keeper
				},
			}

			server = NewServer(app, log.NewNoopLogger())
		)

		result, err := server.VMFile(nil, expectedHeight, expectedPath)
		require.NoError(t, err)

		assert.Equal(t, expectedResult, result)
	})
}

func TestServer_VMDoc(t *testing.T) {
	t.Parallel()

	t.Run("invalid context creation", func(t *testing.T) {
		t.Parallel()

		var (
			sdkErr = errors.New("context err")
			app    = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					return sdk.Context{}, sdkErr
				},
			}

			server = NewServer(app, log.NewNoopLogger())

			height = int64(0)
			pkg    = "gno.land/r/example"
		)

		result, err := server.VMDoc(nil, height, pkg)
		require.Empty(t, result)

		assert.ErrorIs(t, err, sdkErr)
	})

	t.Run("invalid doc query", func(t *testing.T) {
		t.Parallel()

		var (
			queryErr       = errors.New("query err")
			expectedPkg    = "gno.land/r/example"
			expectedHeight = int64(0)

			keeper = &mock.VMKeeper{
				QueryDocFn: func(_ sdk.Context, pkgPath string) (*doc.JSONDocumentation, error) {
					require.Equal(t, expectedPkg, pkgPath)

					return nil, queryErr
				},
			}
			app = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					require.Equal(t, expectedHeight, height)

					return sdk.Context{}, nil
				},
				vmKeeperFn: func() vm.VMKeeperI {
					return keeper
				},
			}

			server = NewServer(app, log.NewNoopLogger())
		)

		result, err := server.VMDoc(nil, expectedHeight, expectedPkg)
		require.Empty(t, result)

		assert.ErrorIs(t, err, queryErr)
	})

	t.Run("valid doc query", func(t *testing.T) {
		t.Parallel()

		var (
			expectedPkg    = "gno.land/r/example"
			expectedHeight = int64(0)
			expectedDoc    = &doc.JSONDocumentation{}
			expectedJSON   = expectedDoc.JSON()

			keeper = &mock.VMKeeper{
				QueryDocFn: func(_ sdk.Context, pkgPath string) (*doc.JSONDocumentation, error) {
					require.Equal(t, expectedPkg, pkgPath)

					return expectedDoc, nil
				},
			}
			app = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					require.Equal(t, expectedHeight, height)

					return sdk.Context{}, nil
				},
				vmKeeperFn: func() vm.VMKeeperI {
					return keeper
				},
			}

			server = NewServer(app, log.NewNoopLogger())
		)

		result, err := server.VMDoc(nil, expectedHeight, expectedPkg)
		require.NoError(t, err)

		assert.Equal(t, expectedJSON, result)
	})
}

func TestServer_VMStorage(t *testing.T) {
	t.Parallel()

	t.Run("invalid context creation", func(t *testing.T) {
		t.Parallel()

		var (
			sdkErr = errors.New("context err")
			app    = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					return sdk.Context{}, sdkErr
				},
			}

			server = NewServer(app, log.NewNoopLogger())

			height = int64(0)
			pkg    = "gno.land/r/example"
		)

		result, err := server.VMStorage(nil, height, pkg)
		require.Empty(t, result)

		assert.ErrorIs(t, err, sdkErr)
	})

	t.Run("invalid storage query", func(t *testing.T) {
		t.Parallel()

		var (
			queryErr       = errors.New("query err")
			expectedPkg    = "gno.land/r/example"
			expectedHeight = int64(0)

			keeper = &mock.VMKeeper{
				QueryStorageFn: func(_ sdk.Context, pkgPath string) (string, error) {
					require.Equal(t, expectedPkg, pkgPath)

					return "", queryErr
				},
			}
			app = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					require.Equal(t, expectedHeight, height)

					return sdk.Context{}, nil
				},
				vmKeeperFn: func() vm.VMKeeperI {
					return keeper
				},
			}

			server = NewServer(app, log.NewNoopLogger())
		)

		result, err := server.VMStorage(nil, expectedHeight, expectedPkg)
		require.Empty(t, result)

		assert.ErrorIs(t, err, queryErr)
	})

	t.Run("valid storage query", func(t *testing.T) {
		t.Parallel()

		var (
			expectedPkg     = "gno.land/r/example"
			expectedHeight  = int64(0)
			expectedStorage = "storage: 10, deposit: 100"

			keeper = &mock.VMKeeper{
				QueryStorageFn: func(_ sdk.Context, pkgPath string) (string, error) {
					require.Equal(t, expectedPkg, pkgPath)

					return expectedStorage, nil
				},
			}
			app = &mockApplication{
				newQueryContextFn: func(height int64) (sdk.Context, error) {
					require.Equal(t, expectedHeight, height)

					return sdk.Context{}, nil
				},
				vmKeeperFn: func() vm.VMKeeperI {
					return keeper
				},
			}

			server = NewServer(app, log.NewNoopLogger())
		)

		result, err := server.VMStorage(nil, expectedHeight, expectedPkg)
		require.NoError(t, err)

		assert.Equal(t, expectedStorage, result)
	})
}
