package vm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParamsRestricted(t *testing.T) {
	env := setupTestEnv()
	params := NewSDKParams(env.vmk, env.ctx)
	params.SetCurRealmPath("gno.land/r/foo")

	testCases := []struct {
		name        string
		setFunc     func()
		expectedMsg string
	}{
		{
			name: "SetString should panic",
			setFunc: func() {
				params.SetString("name.string", "foo")
			},
			expectedMsg: "Set parameters can only be accessed from: " + ParamsRealmPath,
		},
		{
			name: "SetBool should panic",
			setFunc: func() {
				params.SetBool("isFoo.bool", true)
			},
			expectedMsg: "Set parameters can only be accessed from: " + ParamsRealmPath,
		},
		{
			name: "SetInt64 should panic",
			setFunc: func() {
				params.SetInt64("nummber.int64", -100)
			},
			expectedMsg: "Set parameters can only be accessed from: " + ParamsRealmPath,
		},
		{
			name: "SetUint64 should panic",
			setFunc: func() {
				params.SetUint64("nummber.uint64", 100)
			},
			expectedMsg: "Set parameters can only be accessed from: " + ParamsRealmPath,
		},
		{
			name: "SetBytes should panic",
			setFunc: func() {
				params.SetBytes("name.bytes", []byte("foo"))
			},
			expectedMsg: "Set parameters can only be accessed from: " + ParamsRealmPath,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.PanicsWithValue(t, tc.expectedMsg, tc.setFunc, "The panic message did not match the expected value")
		})
	}
}
