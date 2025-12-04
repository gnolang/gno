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
				params.SetString("vm:p", "foo")

				var actual string
				params.pmk.GetString(env.ctx, "vm:p", &actual)
				require.Equal(t, "foo", actual)
			},
		},
		{
			name: "int64 test module vm",
			testFunc: func() {
				params.SetInt64("vm:p", int64(1))

				var actual int64
				params.pmk.GetInt64(env.ctx, "vm:p", &actual)
				require.Equal(t, int64(1), actual)
			},
		},
		{
			name: "string test module auth",
			testFunc: func() {
				params.SetString("auth:p1", "foo")

				var actual string
				params.pmk.GetString(env.ctx, "auth:p1", &actual)
				require.Equal(t, "foo", actual)
			},
		},

		{
			name: "string test module bank",
			testFunc: func() {
				params.SetString("bank:p1", "foo")

				var actual string
				params.pmk.GetString(env.ctx, "auth:p1", &actual)
				require.Equal(t, "foo", actual)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.testFunc()
		})
	}
}
