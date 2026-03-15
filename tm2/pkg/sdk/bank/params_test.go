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

func TestNewParams(t *testing.T) {
	denoms := []string{"ugnot", "atom"}
	p := NewParams(denoms)
	assert.Equal(t, denoms, p.RestrictedDenoms)
}

func TestDefaultParams(t *testing.T) {
	p := DefaultParams()
	assert.Empty(t, p.RestrictedDenoms)
}

func TestParamsString(t *testing.T) {
	p := NewParams([]string{"ugnot"})
	s := p.String()
	assert.Contains(t, s, "Params:")
	assert.Contains(t, s, "ugnot")
}

func TestSetParams(t *testing.T) {
	env := setupTestEnv()

	// valid
	err := env.bankk.SetParams(env.ctx, NewParams([]string{"ugnot"}))
	require.NoError(t, err)
	got := env.bankk.GetParams(env.ctx)
	assert.Equal(t, []string{"ugnot"}, got.RestrictedDenoms)

	// invalid denom
	err = env.bankk.SetParams(env.ctx, NewParams([]string{"!!"}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid restricted denom")
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
