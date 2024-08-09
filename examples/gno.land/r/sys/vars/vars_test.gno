package vars

import (
	"sort"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

func TestVars_GetValue(t *testing.T) {
	t.Parallel()

	t.Run("missing value", func(t *testing.T) {
		t.Parallel()

		uassert.PanicsWithMessage(t, ErrMissingValue.Error(), func() {
			GetValue("random key")
		})
	})

	t.Run("existing value", func(t *testing.T) {
		t.Parallel()

		vars = avl.NewTree()

		var (
			key   = "new key"
			value = "new value"
		)

		vars.Set(key, value)

		uassert.NotPanics(t, func() {
			valRaw := GetValue(key)

			val, ok := valRaw.(string)
			urequire.True(t, ok)

			uassert.Equal(t, value, val)
		})
	})
}

func TestVars_GetStringValue(t *testing.T) {
	t.Parallel()

	t.Run("missing value", func(t *testing.T) {
		t.Parallel()

		uassert.PanicsWithMessage(t, ErrMissingValue.Error(), func() {
			GetStringValue("random key")
		})
	})

	t.Run("existing string value", func(t *testing.T) {
		t.Parallel()

		vars = avl.NewTree()

		var (
			key   = "new string key"
			value = "new value"
		)

		vars.Set(key, value)

		uassert.NotPanics(t, func() {
			val := GetStringValue(key)

			uassert.Equal(t, value, val)
		})
	})
}

func TestVars_GetKeys(t *testing.T) {
	t.Parallel()

	t.Run("no keys", func(t *testing.T) {
		t.Parallel()

		vars = avl.NewTree()

		keys := GetKeys(0, maxRequestKeys)

		// Make sure there are 0 keys
		urequire.Equal(t, 0, len(keys))
	})

	t.Run("existing keys", func(t *testing.T) {
		t.Parallel()

		vars = avl.NewTree()

		savedKeys := make([]string, 0, maxRequestKeys)
		for i := 0; i < maxRequestKeys; i++ {
			kv := KV{
				Key:   ufmt.Sprintf("key-%d", i),
				Value: ufmt.Sprintf("value-%d", i),
			}

			savedKeys = append(savedKeys, kv.Key)

			vars.Set(kv.Key, kv.Value)
		}

		// Fetch the keys
		keys := GetKeys(0, maxRequestKeys)
		urequire.Equal(t, maxRequestKeys, len(keys))

		sort.Strings(keys)
		sort.Strings(savedKeys)

		for index, key := range keys {
			uassert.Equal(t, savedKeys[index], key)
		}
	})
}
