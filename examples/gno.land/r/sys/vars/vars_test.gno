package vars

import (
	"testing"

	"gno.land/p/demo/uassert"
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

		uassert.NotPanics(t, func() {
			_ = GetValue(ValidatorsRealm)
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

		uassert.NotPanics(t, func() {
			_ = GetStringValue(ValidatorsRealm)
		})
	})
}

func TestVars_GetKeys(t *testing.T) {
	t.Parallel()

	keys := GetKeys(0, maxRequestKeys)

	// Make sure there is 1 key
	urequire.Equal(t, 1, len(keys))
	uassert.Equal(t, ValidatorsRealm, keys[0])
}
