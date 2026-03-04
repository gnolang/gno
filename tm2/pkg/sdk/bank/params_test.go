package bank

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWillSetParam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setValue    func(env testEnv)
		shouldPanic bool
	}{
		{
			name: "valid restricted_denoms",
			setValue: func(env testEnv) {
				env.prmk.SetStrings(env.ctx, "bank:p:restricted_denoms", []string{"ugnot"})
			},
			shouldPanic: false,
		},
		{
			name: "wrong type for restricted_denoms",
			setValue: func(env testEnv) {
				env.prmk.SetString(env.ctx, "bank:p:restricted_denoms", "ugnot")
			},
			shouldPanic: true,
		},
		{
			name: "invalid denom for restricted_denoms",
			setValue: func(env testEnv) {
				env.prmk.SetStrings(env.ctx, "bank:p:restricted_denoms", []string{"!!"})
			},
			shouldPanic: true,
		},
		{
			name: "unknown param key panics",
			setValue: func(env testEnv) {
				env.prmk.SetString(env.ctx, "bank:p:nonexistent", "foo")
			},
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			env := setupTestEnv()
			if tt.shouldPanic {
				require.Panics(t, func() { tt.setValue(env) })
			} else {
				require.NotPanics(t, func() { tt.setValue(env) })
			}
		})
	}
}
