package flippandoserver

import (
	"sort"
	"std"
	"testing"
	"time"
	"encoding/json"
	"strconv"
	"errors"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
)

func TestCreateGameBoard(t *testing.T) {
    tileType := "testTileType"
    boardSize := 16  // Assuming a 4x4 board

    //expectedGameTilesLength := int(math.Sqrt(float64(boardSize)))
	expectedGameTilesLength := int(sqrt(float64(boardSize)))

    board, gameTiles, boardSize := CreateGameBoard(tileType, boardSize)
	
}

func TestGenerateRandomNumbers(t *testing.T) {
	tests := []struct {
		howMany int
		min     int
		max     int
		seed    int64
		expect  int
	}{
		{5, 1, 10, 123, 5},
		{3, 5, 7, 456, 3},
		{10, 1, 5, 789, 0},
		{0, 1, 10, 101, 0},
		{-2, 1, 10, 112, 0},
		{5, 10, 5, 123, 0},
	}

	for _, tt := range tests {
		got := generateRandomNumbers(tt.howMany, tt.min, tt.max, tt.seed)

		if len(got) != tt.expect {
			t.Errorf("expected %d numbers but got %d", tt.expect, len(got))
		}

		// Check for uniqueness
		numSet := make(map[int]bool)
		for _, num := range got {
			if numSet[num] {
				t.Errorf("duplicate number found: %d", num)
			}
			numSet[num] = true
		}

		// Check if numbers are within range
		for _, num := range got {
			if num < tt.min || num > tt.max {
				t.Errorf("number %d out of range [%d, %d]", num, tt.min, tt.max)
			}
		}
	}
}


// Helper function to compare two slices for equality
func sliceEqual(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
