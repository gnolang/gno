package vm

import (
	"testing"

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
