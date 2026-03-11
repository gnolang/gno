package bank

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestWillSetParamExhaustive ensures every Params field has a WillSetParam case.
func TestWillSetParamExhaustive(t *testing.T) {
	env := setupTestEnv()

	validValues := map[string]any{
		"restricted_denoms": []string{"ugnot"},
	}

	typ := reflect.TypeOf(Params{})
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		value, ok := validValues[jsonTag]
		require.True(t, ok,
			"field %s (param key p:%s) has no WillSetParam handler and is then unsettable",
			field.Name, jsonTag)

		t.Run(jsonTag, func(t *testing.T) {
			require.NotPanics(t, func() {
				env.bankk.WillSetParam(env.ctx, "p:"+jsonTag, value)
			})
		})
	}
}

func TestWillSetParam(t *testing.T) {
	env := setupTestEnv()

	tests := []struct {
		name        string
		setValue    func()
		shouldPanic bool
	}{
		{
			name: "valid restricted_denoms",
			setValue: func() {
				env.prmk.SetStrings(env.ctx, "bank:p:restricted_denoms", []string{"ugnot"})
			},
			shouldPanic: false,
		},
		{
			name: "wrong type for restricted_denoms",
			setValue: func() {
				env.prmk.SetString(env.ctx, "bank:p:restricted_denoms", "ugnot")
			},
			shouldPanic: true,
		},
		{
			name: "invalid denom for restricted_denoms",
			setValue: func() {
				env.prmk.SetStrings(env.ctx, "bank:p:restricted_denoms", []string{"!!"})
			},
			shouldPanic: true,
		},
		{
			name: "unknown param key panics",
			setValue: func() {
				env.prmk.SetString(env.ctx, "bank:p:nonexistent", "foo")
			},
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				require.Panics(t, tt.setValue)
			} else {
				require.NotPanics(t, tt.setValue)
			}
		})
	}
}
