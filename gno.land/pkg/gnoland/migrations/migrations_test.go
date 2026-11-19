package migrations

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func noop(sdk.Context, Env) error { return nil }

func TestFind(t *testing.T) {
	t.Parallel()

	upgrades := []Upgrade{{Version: "v1.8.0", Handler: noop}, {Version: "v1.9.0", Handler: noop}}

	u, ok := Find(upgrades, "v1.9.0")
	require.True(t, ok)
	assert.Equal(t, "v1.9.0", u.Version)

	for _, v := range []string{"", "v1.10.0"} {
		_, ok := Find(upgrades, v)
		assert.False(t, ok, "version %q", v)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	require.NoError(t, Validate(nil))
	require.NoError(t, Validate([]Upgrade{{Version: "v1.8.0", Handler: noop}, {Version: "v1.9.0", Handler: noop}}))

	for name, upgrades := range map[string][]Upgrade{
		"empty version":     {{Handler: noop}},
		"nil handler":       {{Version: "v1.8.0"}},
		"duplicate version": {{Version: "v1.8.0", Handler: noop}, {Version: "v1.8.0", Handler: noop}},
	} {
		assert.Error(t, Validate(upgrades), name)
	}
}
