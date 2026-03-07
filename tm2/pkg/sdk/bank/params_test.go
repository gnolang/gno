package bank

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
