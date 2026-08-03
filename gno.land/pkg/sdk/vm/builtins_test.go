package vm

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"

	"github.com/stretchr/testify/require"
)

func TestParamsKeeperFail(t *testing.T) {
	env := setupTestEnv()
	params := NewSDKParams(env.vmk.prmk, env.ctx)

	testCases := []struct {
		name        string
		setFunc     func()
		expectedMsg string
	}{
		{
			name: "SetString should panic",
			setFunc: func() {
				params.SetString("foo:name", "foo")
			},
			expectedMsg: `module name <foo> not registered`,
		},
		{
			name: "SetString should panic (with realm)",
			setFunc: func() {
				params.SetString("foo:gno.land/r/user/repo:name", "foo")
			},
			expectedMsg: `module name <foo> not registered`,
		},
		{
			name: "SetBool should panic",
			setFunc: func() {
				params.SetBool("foo:name", true)
			},
			expectedMsg: `module name <foo> not registered`,
		},
		{
			name: "SetInt64 should panic",
			setFunc: func() {
				params.SetInt64("foo:name", -100)
			},
			expectedMsg: `module name <foo> not registered`,
		},
		{
			name: "SetUint64 should panic",
			setFunc: func() {
				params.SetUint64("foo:name", 100)
			},
			expectedMsg: `module name <foo> not registered`,
		},
		{
			name: "SetBytes should panic",
			setFunc: func() {
				params.SetBytes("foo:name", []byte("foo"))
			},
			expectedMsg: `module name <foo> not registered`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.PanicsWithValue(t, tc.expectedMsg, tc.setFunc, "The panic message did not match the expected value")
		})
	}
}

func TestParamsKeeperSuccess(t *testing.T) {
	env := setupTestEnv()
	params := NewSDKParams(env.vmk.prmk, env.ctx)

	testCases := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "string test module vm",
			testFunc: func() {
				params.SetString("vm:p:chain_domain", "gno.land")

				var actual string
				params.pmk.GetString(env.ctx, "vm:p:chain_domain", &actual)
				require.Equal(t, "gno.land", actual)
			},
		},
		{
			name: "string test module vm storage_price",
			testFunc: func() {
				params.SetString("vm:p:storage_price", "200ugnot")

				var actual string
				params.pmk.GetString(env.ctx, "vm:p:storage_price", &actual)
				require.Equal(t, "200ugnot", actual)
			},
		},
		{
			name: "int64 test module auth max_memo_bytes",
			testFunc: func() {
				params.SetInt64("auth:p:max_memo_bytes", 512)

				var actual int64
				params.pmk.GetInt64(env.ctx, "auth:p:max_memo_bytes", &actual)
				require.Equal(t, int64(512), actual)
			},
		},
		{
			name: "strings test module bank restricted_denoms",
			testFunc: func() {
				params.SetStrings("bank:p:restricted_denoms", []string{"ugnot"})

				var actual []string
				params.pmk.GetStrings(env.ctx, "bank:p:restricted_denoms", &actual)
				require.Equal(t, []string{"ugnot"}, actual)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.testFunc()
		})
	}
}

// TestAssertIssuable pins the Go-side re-assertion of the realm-denom prefix.
//
// The prefix separates denoms a realm may create from those it may not, so it is
// a security boundary rather than a naming convention: a bare denom reaching
// issuance would let a realm mint the chain's own gas denom. (Storage tier is a
// different question, decided by an allowlist in the bank.) chain/banker's
// assertCoinDenom enforces the full shape, but it lives in interpreted stdlib
// source and is not the last line of defence.
func TestAssertIssuable(t *testing.T) {
	t.Parallel()

	for _, denom := range []string{"ugnot", "atom", "foo-bar", "gno.land/r/demo/foo:gold"} {
		require.Panics(t, func() { assertIssuable(denom) },
			"a genesis-tier denom must not be issuable by a realm: %s", denom)
	}
	for _, denom := range []string{
		"/gno.land/r/demo/foo:gold",
		"/gno.land/r/my-org/token:gold",
	} {
		require.NotPanics(t, func() { assertIssuable(denom) },
			"a realm-qualified denom must be issuable: %s", denom)
	}
}

// TestSDKBankerRejectsGenesisDenom pins the assertIssuable *call sites*, not just
// the function. Testing assertIssuable alone leaves the wiring uncovered: the
// txtar cannot reach it either, because chain/banker's assertCoinDenom enforces a
// strictly stronger condition in .gno and fires first, so the Go-side check never
// runs in any integration test.
func TestSDKBankerRejectsGenesisDenom(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	banker := NewSDKBanker(env.vmk, env.ctx)
	addr := crypto.AddressFromPreimage([]byte("holder"))

	require.Panics(t, func() { banker.IssueCoin(crypto.Bech32Address(addr.String()), "ugnot", 1) },
		"IssueCoin must refuse a denom that is not realm-qualified")
	require.Panics(t, func() { banker.RemoveCoin(crypto.Bech32Address(addr.String()), "ugnot", 1) },
		"RemoveCoin must refuse a denom that is not realm-qualified")
}
