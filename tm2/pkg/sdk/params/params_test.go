package params

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/require"
)

// newModuleParams defines the parameters with updated fields for a module.
type newModuleParams struct {
	LimitedTokens []string `json:"limited_tokens" yaml:"limited_tokens"`
	Max           uint64   `json:"max" yaml:"max"`
}

// oldModuleParams defines the parameters for a module.
type oldModuleParams struct {
	LimitedTokens []string `json:"limited_tokens" yaml:"limited_tokens"`
}

// newOldModuleParams creates a new oldModuleParams object.
func newOldModuleParams(tokens []string) oldModuleParams {
	return oldModuleParams{
		LimitedTokens: tokens,
	}
}

func TestBackwardCompatibility(t *testing.T) {
	oldParams := newOldModuleParams([]string{"token1", "token2"})

	// Serialize oldModuleParams to JSON
	bz, err := amino.MarshalJSON(oldParams)
	require.NoError(t, err, "Failed to marshal oldModuleParams")

	t.Logf("Serialized oldModuleParams: %s\n", bz)

	// Deserialize JSON into newModuleParams
	newParams := &newModuleParams{}
	err = amino.UnmarshalJSON(bz, newParams)
	require.NoError(t, err, "Failed to unmarshal into newModuleParams")

	// Validate compatibility
	require.Equal(t, oldParams.LimitedTokens, newParams.LimitedTokens, "LimitedTokens mismatch")
	require.Equal(t, uint64(0), newParams.Max, "Max should default to 0 in backward compatibility")
}

func TestForwardCompatibility(t *testing.T) {
	newParams := newModuleParams{LimitedTokens: []string{"token1", "token2"}, Max: 10}

	// Serialize newModuleParams to JSON
	bz, err := amino.MarshalJSON(newParams)
	require.NoError(t, err, "Failed to marshal newModuleParams")

	t.Logf("Serialized newModuleParams: %s\n", bz)

	// Deserialize JSON into oldModuleParams
	oldParams := &oldModuleParams{}
	err = amino.UnmarshalJSON(bz, oldParams)
	require.NoError(t, err, "Failed to unmarshal into oldModuleParams")

	// Validate compatibility
	require.Equal(t, newParams.LimitedTokens, oldParams.LimitedTokens, "LimitedTokens mismatch")
}
