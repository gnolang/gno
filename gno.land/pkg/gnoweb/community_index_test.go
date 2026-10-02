package gnoweb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommunityIndexText(t *testing.T) {
	t.Parallel()

	for _, want := range []CommunityIndex{IndexNoCommunity, IndexRegisteredCommunity, IndexAllCommunity} {
		t.Run(want.String(), func(t *testing.T) {
			t.Parallel()

			text, err := want.MarshalText()
			require.NoError(t, err)
			var got CommunityIndex
			require.NoError(t, got.UnmarshalText(text))
			assert.Equal(t, want, got)
		})
	}

	t.Run("unknown value", func(t *testing.T) {
		t.Parallel()

		var got CommunityIndex
		assert.Error(t, got.UnmarshalText([]byte("some")))
	})
}
