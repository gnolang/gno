package bank

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWillSetParamExhaustive ensures every Params field has a WillSetParam case.
func TestWillSetParamExhaustive(t *testing.T) {
	env := setupTestEnv()

	call := func(param string) (pnc any) {
		defer func() {
			pnc = recover()
		}()
		env.bankk.WillSetParam(env.ctx, param, "")
		return nil
	}

	// baseline: ensure a non-existent key has the expected error.
	const format = "unknown bank param key: %q"
	assert.Equal(t, fmt.Sprintf(format, "doesnotexist"), call("doesnotexist"))

	typ := reflect.TypeOf(Params{})
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		jsonTag, _, _ := strings.Cut(field.Tag.Get("json"), ",")

		t.Run(jsonTag, func(t *testing.T) {
			assert.NotEqual(t, fmt.Sprintf(format, "p:"+jsonTag), call("p:"+jsonTag))
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
