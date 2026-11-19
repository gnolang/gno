package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

// queryCtx cache-wraps the test context as BaseApp wraps a query's: iavl
// iterators charge no gas themselves, only the cache store over them does, so
// on the bare test store vm/qpaths and vm/qinertpaths would run unmetered.
func queryCtx(env testEnv) sdk.Context {
	return env.ctx.WithMultiStore(env.ctx.MultiStore().MultiCacheWrap())
}

// everyQueryEndpoint lists one query per vm endpoint against spinRealmSource
// deployed at pkgPath.
func everyQueryEndpoint(t *testing.T, env testEnv, pkgPath string) []struct{ path, data string } {
	t.Helper()
	listJSON, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetList()")
	require.NoError(t, err)
	oid := extractNestedRefValueObjectID(t, listJSON)
	require.NotEmpty(t, oid)
	return []struct{ path, data string }{
		{"vm/qrender", pkgPath + ":"},
		{"vm/qfuncs", pkgPath},
		{"vm/qeval", pkgPath + ".List"},
		{"vm/qeval_json", pkgPath + ".List"},
		{"vm/qobject_json", oid},
		{"vm/qobject_binary", oid},
		{"vm/qfile", pkgPath},
		{"vm/qdoc", pkgPath},
		{"vm/qpaths", "gno.land/r/"},
		{"vm/qstorage", pkgPath},
		{"vm/qpkg_json", pkgPath},
		{"vm/qpkgmeta_json", pkgPath},
		{"vm/qinertpaths", ""},
		{"vm/qtype_json", pkgPath + ".T"},
	}
}

// TestVmHandlerQuery_OutOfGasEveryEndpoint runs every vm query endpoint at a gas
// budget of 1. Each must answer an out-of-gas error, not panic out of the
// handler: several run out before any machine exists (loading the package, the
// object, the type), where only a doRecoverQueryNoMachine defer reports it.
func TestVmHandlerQuery_OutOfGasEveryEndpoint(t *testing.T) {
	env := setupSizeGuardRealm(t, "spinaddr", spinRealmPath, spinRealmSource)
	endpoints := everyQueryEndpoint(t, env, spinRealmPath)

	qctx := queryCtx(env)
	env.vmk.QueryMaxGas = 1
	for _, ep := range endpoints {
		var res abci.ResponseQuery
		require.NotPanics(t, func() {
			res = env.vmh.Query(qctx, abci.RequestQuery{Path: ep.path, Data: []byte(ep.data)})
		}, ep.path)
		require.Error(t, res.Error, ep.path)
		assert.Contains(t, res.Log, "OutOfGasError", ep.path)
	}
}
