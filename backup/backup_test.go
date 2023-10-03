package backup

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackup_DetermineRightBound(t *testing.T) {
	t.Parallel()

	t.Run("unable to fetch latest block number", func(t *testing.T) {
		t.Parallel()

		var (
			fetchErr   = errors.New("unable to fetch latest height")
			mockClient = &mockClient{
				getLatestBlockNumberFn: func() (uint64, error) {
					return 0, fetchErr
				},
			}
		)

		// Determine the right bound
		_, err := determineRightBound(mockClient, nil)

		assert.ErrorIs(t, err, fetchErr)
	})

	t.Run("excessive right range", func(t *testing.T) {
		t.Parallel()

		var (
			chainLatest uint64 = 10
			requestedTo        = chainLatest + 10 // > chain latest

			mockClient = &mockClient{
				getLatestBlockNumberFn: func() (uint64, error) {
					return chainLatest, nil
				},
			}
		)

		// Determine the right bound
		rightBound, err := determineRightBound(mockClient, &requestedTo)
		require.NoError(t, err)

		assert.Equal(t, chainLatest, rightBound)
	})

	t.Run("valid right range", func(t *testing.T) {
		t.Parallel()

		var (
			chainLatest uint64 = 10
			requestedTo        = chainLatest / 2 // < chain latest

			mockClient = &mockClient{
				getLatestBlockNumberFn: func() (uint64, error) {
					return chainLatest, nil
				},
			}
		)

		// Determine the right bound
		rightBound, err := determineRightBound(mockClient, &requestedTo)
		require.NoError(t, err)

		assert.Equal(t, requestedTo, rightBound)
	})
}
