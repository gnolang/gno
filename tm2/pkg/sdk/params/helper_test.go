package params

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustParamString(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		got := MustParamString("foo", "bar")
		require.Equal(t, "bar", got)
	})

	t.Run("wrong type panics", func(t *testing.T) {
		assert.PanicsWithValue(t,
			"invalid type for foo param: expected string, got int",
			func() { MustParamString("foo", 42) },
		)
	})
}

func TestMustParamInt64(t *testing.T) {
	t.Run("valid int64", func(t *testing.T) {
		got := MustParamInt64("num", int64(99))
		require.Equal(t, int64(99), got)
	})

	t.Run("wrong type panics", func(t *testing.T) {
		assert.PanicsWithValue(t,
			"invalid type for num param: expected int64, got string",
			func() { MustParamInt64("num", "not a number") },
		)
	})
}

func TestMustParamStrings(t *testing.T) {
	t.Run("valid []string", func(t *testing.T) {
		got := MustParamStrings("tags", []string{"a", "b"})
		require.Equal(t, []string{"a", "b"}, got)
	})

	t.Run("wrong type panics", func(t *testing.T) {
		assert.PanicsWithValue(t,
			"invalid type for tags param: expected []string, got string",
			func() { MustParamStrings("tags", "not a slice") },
		)
	})
}
