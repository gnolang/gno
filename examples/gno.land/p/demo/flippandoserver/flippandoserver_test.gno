package flippandoserver

import (
	"sort"
	"math"
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
	tests := []struct {
		name       string
		tileType   string
		boardSize  int
		seed       int64
		maxTileVal int
	}{
		{"SquareGrid", "squareGrid", 4, 12345, 16},
		{"Hexagrams", "hexagrams", 4, 12345, 4},
		{"Dice", "dice", 4, 12345, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, gameTiles, size := CreateGameBoard(tt.tileType, tt.boardSize, tt.seed)

			// Test the length of the board
			if len(board) != tt.boardSize {
				t.Errorf("Expected board length of %d, got %d", tt.boardSize, len(board))
			}

			// Test the length of gameTiles
			expectedGameTilesLength := int(sqrt(float64(tt.boardSize)))
			if len(gameTiles) != expectedGameTilesLength {
				t.Errorf("Expected gameTiles length of %d, got %d", expectedGameTilesLength, len(gameTiles))
			}

			// Check if gameTiles values are within the expected range
			for _, tile := range gameTiles {
				if tile < 1 || tile > int64(tt.maxTileVal) {
					t.Errorf("gameTiles contains a value out of expected range (1 to %d): got %d", tt.maxTileVal, tile)
				}
			}

			// Check if the gameTiles values are unique
			uniqueTiles := make(map[int64]bool)
			for _, tile := range gameTiles {
				if uniqueTiles[tile] {
					t.Errorf("Duplicate tile found in gameTiles: %d", tile)
				}
				uniqueTiles[tile] = true
			}

			// Test the returned board size
			if size != tt.boardSize {
				t.Errorf("Expected returned board size to be %d, got %d", tt.boardSize, size)
			}
		})
	}
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
		got := GenerateRandomNumbers(tt.howMany, tt.min, tt.max, tt.seed)

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
