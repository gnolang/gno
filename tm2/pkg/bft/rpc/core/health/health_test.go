package health

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_Health(t *testing.T) {
	t.Parallel()

	t.Run("Unexpected params", func(t *testing.T) {
		t.Parallel()

		res, err := HealthHandler(nil, []any{"extra"})
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Valid health status", func(t *testing.T) {
		t.Parallel()

		res, err := HealthHandler(nil, nil)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultHealth)
		require.True(t, ok)

		assert.Equal(t, &ResultHealth{}, result)
	})
}
