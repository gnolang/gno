package utils

import (
	"errors"
	"fmt"
)

// NormalizeHeight normalizes a requested height against the current chain height.
//
// Semantics:
//   - requestedHeight == 0 -> use latest height
//   - requestedHeight < minVal -> error
//   - requestedHeight > currentHeight -> error
func NormalizeHeight(latestHeight, requestedHeight, minVal int64) (int64, error) {
	// 0 means unspecified -> latest
	if requestedHeight == 0 {
		return latestHeight, nil
	}

	if requestedHeight < minVal {
		return 0, fmt.Errorf("height must be greater than or equal to %d", minVal)
	}

	if requestedHeight > latestHeight {
		return 0, errors.New("height must be less than or equal to the current blockchain height")
	}

	return requestedHeight, nil
}
