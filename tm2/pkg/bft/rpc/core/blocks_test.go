package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlockchainInfo(t *testing.T) {
	t.Parallel()

	cases := []struct {
		min, max     int64
		height       int64
		limit        int64
		resultLength int64
		wantErr      bool
	}{
		// min > max
		{0, 0, 0, 10, 0, true},  // min set to 1
		{0, 1, 0, 10, 0, true},  // max set to height (0)
		{0, 0, 1, 10, 1, false}, // max set to height (1)
		{2, 0, 1, 10, 0, true},  // max set to height (1)
		{2, 1, 5, 10, 0, true},

		// negative
		{1, 10, 14, 10, 10, false}, // control
		{-1, 10, 14, 10, 0, true},
		{1, -10, 14, 10, 0, true},
		{-9223372036854775808, -9223372036854775788, 100, 20, 0, true},

		// check limit and height
		{1, 1, 1, 10, 1, false},
		{1, 1, 5, 10, 1, false},
		{2, 2, 5, 10, 1, false},
		{1, 2, 5, 10, 2, false},
		{1, 5, 1, 10, 1, false},
		{1, 5, 10, 10, 5, false},
		{1, 15, 10, 10, 10, false},
		{1, 15, 15, 10, 10, false},
		{1, 15, 15, 20, 15, false},
		{1, 20, 15, 20, 15, false},
		{1, 20, 20, 20, 20, false},
	}

	for i, c := range cases {
		caseString := fmt.Sprintf("test %d failed", i)
		min, max, err := filterMinMax(c.height, c.min, c.max, c.limit)
		if c.wantErr {
			require.Error(t, err, caseString)
		} else {
			require.NoError(t, err, caseString)
			require.Equal(t, 1+max-min, c.resultLength, caseString)
		}
	}
}

func TestGetHeight(t *testing.T) {
	t.Parallel()

	cases := []struct {
		currentHeight int64
		heightPtr     *int64
		min           int64
		res           int64
		wantErr       bool
	}{
		// height >= min
		{42, int64Ptr(0), 0, 0, false},
		{42, int64Ptr(1), 0, 1, false},

		// height < min
		{42, int64Ptr(0), 1, 0, true},

		// nil height
		{42, nil, 1, 42, false},
	}

	for i, c := range cases {
		caseString := fmt.Sprintf("test %d failed", i)
		res, err := getHeightWithMin(c.currentHeight, c.heightPtr, c.min)
		if c.wantErr {
			require.Error(t, err, caseString)
		} else {
			require.NoError(t, err, caseString)
			require.Equal(t, res, c.res, caseString)
		}
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
