package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeHeight(t *testing.T) {
	t.Parallel()

	t.Run("Zero height uses latest", func(t *testing.T) {
		t.Parallel()

		height, err := NormalizeHeight(10, 0, 1)
		require.NoError(t, err)

		assert.Equal(t, int64(10), height)
	})

	t.Run("Below minimum", func(t *testing.T) {
		t.Parallel()

		_, err := NormalizeHeight(10, 1, 2)
		require.Error(t, err)

		assert.Contains(t, err.Error(), "greater than or equal to 2")
	})

	t.Run("Above latest", func(t *testing.T) {
		t.Parallel()

		_, err := NormalizeHeight(10, 11, 1)
		require.Error(t, err)

		assert.Contains(t, err.Error(), "current blockchain height")
	})

	t.Run("Within range", func(t *testing.T) {
		t.Parallel()

		height, err := NormalizeHeight(10, 7, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(7), height)
	})
}
