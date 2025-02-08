package vm

import (
	"testing"

	gstd "github.com/gnolang/gno/gnovm/stdlibs/std"
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
				pk := gstd.ParamKey{
					Realm:  "gno.land/p/foo",
					Prefix: "bank",
					Key:    "name",
					Type:   "string",
				}
				params.SetString(pk, "foo")
			},
			expectedMsg: "Set parameters must be accessed from a realm",
		},
		{
			name: "SetBool should panic",
			setFunc: func() {
				pk := gstd.ParamKey{
					Realm:  "gno.land/p/foo",
					Prefix: "bank",
					Key:    "isFoo",
					Type:   "bool",
				}
				params.SetBool(pk, true)
			},
			expectedMsg: "Set parameters must be accessed from a realm",
		},
		{
			name: "SetInt64 should panic",
			setFunc: func() {
				pk := gstd.ParamKey{
					Realm:  "gno.land/p/foo",
					Prefix: "bank",
					Key:    "number",
					Type:   "int64",
				}
				params.SetInt64(pk, -100)
			},
			expectedMsg: "Set parameters must be accessed from a realm",
		},
		{
			name: "SetUint64 should panic",
			setFunc: func() {
				pk := gstd.ParamKey{
					Realm:  "gno.land/p/foo",
					Prefix: "bank",
					Key:    "number",
					Type:   "uint64",
				}
				params.SetUint64(pk, 100)
			},
			expectedMsg: "Set parameters must be accessed from a realm",
		},
		{
			name: "SetBytes should panic",
			setFunc: func() {
				pk := gstd.ParamKey{
					Realm:  "gno.land/p/foo",
					Prefix: "bank",
					Key:    "name",
					Type:   "bytes",
				}

				params.SetBytes(pk, []byte("foo"))
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
				pk := gstd.ParamKey{
					Realm:  "gno.land/p/foo",
					Prefix: "foo",
					Key:    "name",
					Type:   "string",
				}
				params.SetString(pk, "foo")
			},
			expectedMsg: `keeper key <foo> does not exist`,
		},
		{
			name: "SetBool should panic",
			setFunc: func() {
				pk := gstd.ParamKey{
					Realm:  "gno.land/p/foo",
					Prefix: "foo",
					Key:    "isFoo",
					Type:   "bool",
				}
				params.SetBool(pk, true)
			},
			expectedMsg: `keeper key <foo> does not exist`,
		},
		{
			name: "SetInt64 should panic",
			setFunc: func() {
				pk := gstd.ParamKey{
					Realm:  "gno.land/p/foo",
					Prefix: "foo",
					Key:    "number",
					Type:   "int64",
				}
				params.SetInt64(pk, -100)
			},
			expectedMsg: `keeper key <foo> does not exist`,
		},
		{
			name: "SetUint64 should panic",
			setFunc: func() {
				pk := gstd.ParamKey{
					Realm:  "gno.land/p/foo",
					Prefix: "foo",
					Key:    "number",
					Type:   "uint64",
				}
				params.SetUint64(pk, 100)
			},
			expectedMsg: `keeper key <foo> does not exist`,
		},
		{
			name: "SetBytes should panic",
			setFunc: func() {
				pk := gstd.ParamKey{
					Realm:  "gno.land/p/foo",
					Prefix: "foo",
					Key:    "name",
					Type:   "bytes",
				}
				params.SetBytes(pk, []byte("foo"))
			},
			expectedMsg: `keeper key <foo> does not exist`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.PanicsWithValue(t, tc.expectedMsg, tc.setFunc, "The panic message did not match the expected value")
		})
	}
}
