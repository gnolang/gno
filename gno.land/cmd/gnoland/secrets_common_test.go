package main

import (
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommon_SaveReadNodeKey(t *testing.T) {
	t.Parallel()

	t.Run("invalid data save path", func(t *testing.T) {
		t.Parallel()

		assert.ErrorContains(
			t,
			saveNodeKey(nil, ""),
			"unable to save data to path",
		)
	})

	t.Run("invalid data read path", func(t *testing.T) {
		t.Parallel()

		readData, err := readNodeKey("")
		assert.Nil(t, readData)

		assert.ErrorContains(
			t,
			err,
			"unable to read data",
		)
	})

	t.Run("valid data save and read", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "key.json")
		key := types.GenerateNodeKey()

		require.NoError(t, saveNodeKey(key, path))

		readKey, err := readNodeKey(path)
		require.NoError(t, err)

		assert.Equal(t, key, readKey)
	})
}

func TestCommon_ValidateNodeKey(t *testing.T) {
	t.Parallel()

	t.Run("valid node key", func(t *testing.T) {
		t.Parallel()

		key := types.GenerateNodeKey()

		assert.NoError(t, validateNodeKey(key))
	})

	t.Run("invalid node key", func(t *testing.T) {
		t.Parallel()

		key := types.GenerateNodeKey()
		key.PrivKey = nil

		assert.ErrorIs(t, validateNodeKey(key), errInvalidNodeKey)
	})
}
