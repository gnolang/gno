package launchpad

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestBlockTimeInfo_BlockTime(t *testing.T) {
	tests := []struct {
		name      string
		blockTime int64
		expected  int64
	}{
		{
			name:      "get block time successfully",
			blockTime: 1000,
			expected:  1000,
		},
		{
			name:      "get zero block time successfully",
			blockTime: 0,
			expected:  0,
		},
		{
			name:      "get negative block time successfully",
			blockTime: -1000,
			expected:  -1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockTimeInfo := &BlockTimeInfo{
				blockTime: tt.blockTime,
			}
			uassert.Equal(t, tt.expected, blockTimeInfo.BlockTime())
		})
	}
}

func TestBlockTimeInfo_BlockHeight(t *testing.T) {
	tests := []struct {
		name        string
		blockHeight int64
		expected    int64
	}{
		{
			name:        "get block height successfully",
			blockHeight: 1000,
			expected:    1000,
		},
		{
			name:        "get zero block height successfully",
			blockHeight: 0,
			expected:    0,
		},
		{
			name:        "get negative block height successfully",
			blockHeight: -1000,
			expected:    -1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockTimeInfo := &BlockTimeInfo{
				blockHeight: tt.blockHeight,
			}
			uassert.Equal(t, tt.expected, blockTimeInfo.BlockHeight())
		})
	}
}

func TestBlockTimeInfo_UpdateHeightByAverageBlockTimeMs(t *testing.T) {
	tests := []struct {
		name                 string
		blockTime            int64
		currentTime          int64
		currentHeight        int64
		averageBlockTimeMs   int64
		expectedHeight       int64
		expectedHasPanic     bool
		expectedPanicMessage string
	}{
		{
			name:               "update height when block time is in future",
			blockTime:          1000,
			currentTime:        900,
			currentHeight:      100,
			averageBlockTimeMs: 2000,
			expectedHeight:     150, // (1000-900)*1000/2000 + 100
		},
		{
			name:               "not update height when block time is in past",
			blockTime:          800,
			currentTime:        900,
			currentHeight:      100,
			averageBlockTimeMs: 2000,
			expectedHeight:     100,
		},
		{
			name:                 "update height occur panic with zero average block time",
			blockTime:            1000,
			currentTime:          900,
			currentHeight:        100,
			averageBlockTimeMs:   0,
			expectedHeight:       100,
			expectedHasPanic:     true,
			expectedPanicMessage: "average block time must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockTimeInfo := &BlockTimeInfo{
				blockTime:   tt.blockTime,
				blockHeight: tt.currentHeight,
			}

			if tt.expectedHasPanic {
				uassert.PanicsWithMessage(t, tt.expectedPanicMessage, func() {
					blockTimeInfo.updateHeightByAverageBlockTimeMs(tt.currentTime, tt.currentHeight, tt.averageBlockTimeMs)
				})
			} else {
				blockTimeInfo.updateHeightByAverageBlockTimeMs(tt.currentTime, tt.currentHeight, tt.averageBlockTimeMs)
				uassert.Equal(t, tt.expectedHeight, blockTimeInfo.BlockHeight())
			}
		})
	}
}

func TestNewBlockTimeInfo(t *testing.T) {
	tests := []struct {
		name           string
		time           int64
		height         int64
		expectedTime   int64
		expectedHeight int64
	}{
		{
			name:           "create new block time info successfully",
			time:           1000,
			height:         100,
			expectedTime:   1000,
			expectedHeight: 100,
		},
		{
			name:           "create new block time info with zero values",
			time:           0,
			height:         0,
			expectedTime:   0,
			expectedHeight: 0,
		},
		{
			name:           "create new block time info with negative values",
			time:           -1000,
			height:         -100,
			expectedTime:   -1000,
			expectedHeight: -100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockTimeInfo := newBlockTimeInfo(tt.time, tt.height)
			uassert.Equal(t, tt.expectedTime, blockTimeInfo.BlockTime())
			uassert.Equal(t, tt.expectedHeight, blockTimeInfo.BlockHeight())
		})
	}
}

func TestNewBlockTimeInfoByTimeAndAverageBlockTimeMs(t *testing.T) {
	tests := []struct {
		name               string
		currentTime        int64
		currentHeight      int64
		targetTime         int64
		averageBlockTimeMs int64
		expectedTime       int64
		expectedHeight     int64
	}{
		{
			name:               "create new block time info with future target time",
			currentTime:        1000,
			currentHeight:      100,
			targetTime:         1200,
			averageBlockTimeMs: 2000,
			expectedTime:       1200,
			expectedHeight:     200, // (1200-1000)*1000/2000 + 100
		},
		{
			name:               "create new block time info with past target time",
			currentTime:        1000,
			currentHeight:      100,
			targetTime:         800,
			averageBlockTimeMs: 2000,
			expectedTime:       800,
			expectedHeight:     0,
		},
		{
			name:               "create new block time info with same target time",
			currentTime:        1000,
			currentHeight:      100,
			targetTime:         1000,
			averageBlockTimeMs: 2000,
			expectedTime:       1000,
			expectedHeight:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockTimeInfo := newBlockTimeInfoByTimeAndAverageBlockTimeMs(
				tt.currentTime,
				tt.currentHeight,
				tt.targetTime,
				tt.averageBlockTimeMs,
			)
			uassert.Equal(t, tt.expectedTime, blockTimeInfo.BlockTime())
			uassert.Equal(t, tt.expectedHeight, blockTimeInfo.BlockHeight())
		})
	}
}
