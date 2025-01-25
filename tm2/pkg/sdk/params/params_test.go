package params

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/require"
)

// NewModuleParams defines the parameters with updated fields for a module.
type NewModuleParams struct {
	LimitedTokens []string `json:"limited_tokens" yaml:"limited_tokens"`
	Max           uint64   `json:"max" yaml:"max"`
}

// OldModuleParams defines the parameters for a module.
type OldModuleParams struct {
	LimitedTokens []string `json:"limited_tokens" yaml:"limited_tokens"`
}

// NewOldModuleParams creates a new OldModuleParams object.
func NewOldModuleParams(tokens []string) OldModuleParams {
	return OldModuleParams{
		LimitedTokens: tokens,
	}
}

func TestBackwardCompatibility(t *testing.T) {
	oldParams := NewOldModuleParams([]string{"token1", "token2"})

	// Serialize OldModuleParams to JSON
	bz, err := amino.MarshalJSON(oldParams)
	require.NoError(t, err, "Failed to marshal OldModuleParams")

	t.Logf("Serialized OldModuleParams: %s\n", bz)

	// Deserialize JSON into NewModuleParams
	newParams := &NewModuleParams{}
	err = amino.UnmarshalJSON(bz, newParams)
	require.NoError(t, err, "Failed to unmarshal into NewModuleParams")

	// Validate compatibility
	require.Equal(t, oldParams.LimitedTokens, newParams.LimitedTokens, "LimitedTokens mismatch")
	require.Equal(t, uint64(0), newParams.Max, "Max should default to 0 in backward compatibility")
}

func TestForwardCompatibility(t *testing.T) {
	newParams := NewModuleParams{LimitedTokens: []string{"token1", "token2"}, Max: 10}

	// Serialize NewModuleParams to JSON
	bz, err := amino.MarshalJSON(newParams)
	require.NoError(t, err, "Failed to marshal NewModuleParams")

	t.Logf("Serialized NewModuleParams: %s\n", bz)

	// Deserialize JSON into OldModuleParams
	oldParams := &OldModuleParams{}
	err = amino.UnmarshalJSON(bz, oldParams)
	require.NoError(t, err, "Failed to unmarshal into OldModuleParams")

	// Validate compatibility
	require.Equal(t, newParams.LimitedTokens, oldParams.LimitedTokens, "LimitedTokens mismatch")
}
