package vm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParamsRestrictedRealm(t *testing.T) {
	env := setupTestEnv()
	params := NewSDKParams(&env.vmk.prmk, env.ctx)

	testCases := []struct {
		name        string
		setFunc     func()
		expectedMsg string
	}{
		{
			name: "SetString should panic",
			setFunc: func() {
				params.SetString("gno.land/p/foo.bank.name.string", "foo")
			},
			expectedMsg: "Set parameters must be accessed from a realm",
		},
		{
			name: "SetBool should panic",
			setFunc: func() {
				params.SetBool("gno.land/p/foo.bank.isFoo.bool", true)
			},
			expectedMsg: "Set parameters must be accessed from a realm",
		},
		{
			name: "SetInt64 should panic",
			setFunc: func() {
				params.SetInt64("gno.land/p/foo.bank.nummber.int64", -100)
			},
			expectedMsg: "Set parameters must be accessed from a realm",
		},
		{
			name: "SetUint64 should panic",
			setFunc: func() {
				params.SetUint64("gno.land/p/foo.bank.nummber.uint64", 100)
			},
			expectedMsg: "Set parameters must be accessed from a realm",
		},
		{
			name: "SetBytes should panic",
			setFunc: func() {
				params.SetBytes("gno.land/p/foo.bank.name.bytes", []byte("foo"))
			},
			expectedMsg: "Set parameters must be accessed from a realm",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.PanicsWithValue(t, tc.expectedMsg, tc.setFunc, "The panic message did not match the expected value")
		})
	}
}

func TestParamsKeeper(t *testing.T) {
	env := setupTestEnv()
	params := NewSDKParams(&env.vmk.prmk, env.ctx)

	testCases := []struct {
		name        string
		setFunc     func()
		expectedMsg string
	}{
		{
			name: "SetString should panic",
			setFunc: func() {
				params.SetString("gno.land/r/sys/params.foo.name.string", "foo")
			},
			expectedMsg: "keeper key foo does not exist",
		},
		{
			name: "SetBool should panic",
			setFunc: func() {
				params.SetBool("gno.land/r/sys/params.foo.isFoo.bool", true)
			},
			expectedMsg: "keeper key foo does not exist",
		},
		{
			name: "SetInt64 should panic",
			setFunc: func() {
				params.SetInt64("gno.land/r/sys/params.foo.nummber.int64", -100)
			},
			expectedMsg: "keeper key foo does not exist",
		},
		{
			name: "SetUint64 should panic",
			setFunc: func() {
				params.SetUint64("gno.land/r/sys/params.foo.nummber.uint64", 100)
			},
			expectedMsg: "keeper key foo does not exist",
		},
		{
			name: "SetBytes should panic",
			setFunc: func() {
				params.SetBytes("gno.land/r/sys/params.foo.name.bytes", []byte("foo"))
			},
			expectedMsg: "keeper key foo does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.PanicsWithValue(t, tc.expectedMsg, tc.setFunc, "The panic message did not match the expected value")
		})
	}
}
