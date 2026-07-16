package vm

import (
	"reflect"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readParamsBundle(t *testing.T, env testEnv, ctx sdk.Context) (Params, bool) {
	t.Helper()

	var bundle []byte
	if !env.prmk.GetBytes(ctx, paramsBundleKey, &bundle) {
		return Params{}, false
	}
	var params Params
	require.NoError(t, amino.Unmarshal(bundle, &params))
	return params, true
}

func TestParamsBundleSetParamsSynchronizesRepresentations(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	params := DefaultParams()
	params.ChainDomain = "example.com"
	params.PreprocessGasPerByte = 2_500
	params.StorageFeeCollector = crypto.AddressFromPreimage([]byte("new-storage-fee-collector"))
	require.NoError(t, env.vmk.SetParams(ctx, params))

	bundled, ok := readParamsBundle(t, env, ctx)
	require.True(t, ok)
	assert.Equal(t, params, bundled)

	var chainDomain string
	require.True(t, env.prmk.GetString(ctx, chainDomainParamPath, &chainDomain))
	assert.Equal(t, params.ChainDomain, chainDomain)

	var preprocessGasPerByte int64
	require.True(t, env.prmk.GetInt64(ctx, "vm:p:preprocess_gas_per_byte", &preprocessGasPerByte))
	assert.Equal(t, params.PreprocessGasPerByte, preprocessGasPerByte)

	var storageFeeCollector string
	require.True(t, env.prmk.GetString(ctx, "vm:p:storage_fee_collector", &storageFeeCollector))
	assert.Equal(t, params.StorageFeeCollector.String(), storageFeeCollector)
}

func TestParamsBundleReadPreferenceAndLegacyFallback(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	bundled := DefaultParams()
	bundled.ChainDomain = "bundle.example"
	bundled.SysNamesPkgPath = "gno.land/r/sys/bundlenames"
	bundled.SysCLAPkgPath = "gno.land/r/sys/bundlecla"
	require.NoError(t, env.vmk.SetParams(ctx, bundled))

	legacy := bundled
	legacy.ChainDomain = "legacy.example"
	legacy.SysNamesPkgPath = "gno.land/r/sys/legacynames"
	legacy.SysCLAPkgPath = "gno.land/r/sys/legacycla"
	env.prmk.SetStruct(ctx, "vm:p", legacy) // bypass SetParams to make the representations diverge

	assert.Equal(t, bundled, env.vmk.GetParams(ctx))
	assert.Equal(t, bundled.ChainDomain, env.vmk.getChainDomainParam(ctx))
	assert.Equal(t, bundled.SysNamesPkgPath, env.vmk.getSysNamesPkgParam(ctx))
	assert.Equal(t, bundled.SysCLAPkgPath, env.vmk.getSysCLAPkgParam(ctx))

	env.prmk.SetBytes(ctx, paramsBundleKey, nil)
	assert.Equal(t, legacy, env.vmk.GetParams(ctx))
	assert.Equal(t, legacy.ChainDomain, env.vmk.getChainDomainParam(ctx))
	assert.Equal(t, legacy.SysNamesPkgPath, env.vmk.getSysNamesPkgParam(ctx))
	assert.Equal(t, legacy.SysCLAPkgPath, env.vmk.getSysCLAPkgParam(ctx))
}

func TestParamsBundleGovernanceSyncAndMigrationGate(t *testing.T) {
	t.Run("active bundle stays synchronized", func(t *testing.T) {
		env := setupTestEnv()
		ctx := env.ctx
		expected := DefaultParams()

		env.prmk.SetString(ctx, chainDomainParamPath, "example.com")
		expected.ChainDomain = "example.com"
		assert.Equal(t, expected, env.vmk.GetParams(ctx))
		bundled, ok := readParamsBundle(t, env, ctx)
		require.True(t, ok)
		assert.Equal(t, expected, bundled)

		env.prmk.SetInt64(ctx, "vm:p:preprocess_gas_per_byte", 2_500)
		expected.PreprocessGasPerByte = 2_500
		assert.Equal(t, expected, env.vmk.GetParams(ctx))
		bundled, ok = readParamsBundle(t, env, ctx)
		require.True(t, ok)
		assert.Equal(t, expected, bundled)

		before, ok := readParamsBundle(t, env, ctx)
		require.True(t, ok)
		assert.Panics(t, func() {
			env.prmk.SetString(ctx, chainDomainParamPath, "not/a/domain")
		})
		after, ok := readParamsBundle(t, env, ctx)
		require.True(t, ok)
		assert.Equal(t, before, after)
		var chainDomain string
		require.True(t, env.prmk.GetString(ctx, chainDomainParamPath, &chainDomain))
		assert.Equal(t, expected.ChainDomain, chainDomain)
	})

	t.Run("legacy governance write does not activate bundle", func(t *testing.T) {
		env := setupTestEnv()
		ctx := env.ctx
		env.prmk.SetBytes(ctx, paramsBundleKey, nil)

		env.prmk.SetString(ctx, chainDomainParamPath, "example.com")
		_, ok := readParamsBundle(t, env, ctx)
		assert.False(t, ok)
		assert.Equal(t, "example.com", env.vmk.GetParams(ctx).ChainDomain)

		params := env.vmk.GetParams(ctx)
		require.NoError(t, env.vmk.SetParams(ctx, params))
		migrated, ok := readParamsBundle(t, env, ctx)
		require.True(t, ok)
		assert.Equal(t, params, migrated)
	})
}

func TestParamsBundleInitGenesisMigratesLegacyState(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	legacy := DefaultParams()
	legacy.ChainDomain = "legacy.example"
	legacy.PreprocessGasPerByte = 0
	env.prmk.SetBytes(ctx, paramsBundleKey, nil)
	env.prmk.SetStruct(ctx, "vm:p", legacy)

	exported := env.vmk.ExportGenesis(ctx)
	assert.Equal(t, preprocessGasPerByteDefault, exported.Params.PreprocessGasPerByte)
	env.vmk.InitGenesis(ctx, exported)

	bundled, ok := readParamsBundle(t, env, ctx)
	require.True(t, ok)
	assert.Equal(t, exported.Params, bundled)
	assert.Equal(t, exported.Params, env.vmk.GetParams(ctx))
}

func TestParamsBundleMalformedStatePanics(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx
	env.prmk.SetBytes(ctx, paramsBundleKey, []byte{0xff})

	assert.Panics(t, func() { env.vmk.GetParams(ctx) })
	assert.Panics(t, func() { env.vmk.getChainDomainParam(ctx) })
}

func TestParamsBundleColdReadGas(t *testing.T) {
	newMeteredContext := func(env testEnv) (sdk.Context, store.GasMeter) {
		cfg := store.DefaultGasConfig()
		cfg.FixedGetReadDepth100 = 100
		cfg.ReadCostPerByte = 0
		meter := store.NewInfiniteGasMeter()
		ctx, _ := env.ctx.WithGasMeter(meter).WithGasConfig(cfg).CacheContext()
		return ctx, meter
	}
	readAllCallSites := func(env testEnv, ctx sdk.Context) {
		env.vmk.GetParams(ctx)
		env.vmk.getChainDomainParam(ctx)
		env.vmk.getSysNamesPkgParam(ctx)
		env.vmk.getSysCLAPkgParam(ctx)
		env.vmk.GetParams(ctx)
	}

	t.Run("bundle", func(t *testing.T) {
		env := setupTestEnv()
		ctx, meter := newMeteredContext(env)
		readAllCallSites(env, ctx)

		assert.Equal(t, store.DefaultGasConfig().ReadCostFlat, meter.GasConsumed(),
			"all VM param call sites must collapse to one cold bundle read")
	})

	t.Run("legacy fallback", func(t *testing.T) {
		env := setupTestEnv()
		env.prmk.SetBytes(env.ctx, paramsBundleKey, nil)
		ctx, meter := newMeteredContext(env)
		readAllCallSites(env, ctx)

		wantReads := int64(1 + reflect.TypeFor[Params]().NumField())
		assert.Equal(t, wantReads*store.DefaultGasConfig().ReadCostFlat, meter.GasConsumed(),
			"legacy fallback reads the absent bundle key plus every individual field once")
	})
}
