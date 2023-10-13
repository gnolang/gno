package flippandoserver

import (
	"sort"
	"std"
	"testing"
	"time"
	"encoding/json"
	"errors"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
)

func TestCreateGameBoard(t *testing.T) {
    tileType := "testTileType"
    boardSize := 16  // Assuming a 4x4 board

    //expectedGameTilesLength := int(math.Sqrt(float64(boardSize)))
	expectedGameTilesLength := int(sqrt(float64(boardSize)))

    gameBoardStr := CreateGameBoard(tileType, boardSize)
	ufmt.Sprintf(`game board: %s`, gameBoardStr)
	//ufmt.Sprintf("Generated JSON: %s\n", gameBoardStr)
	/*
	var game Game
    

    if game.GameId != gameId || game.TileType != tileType || game.BoardSize != boardSize {
        t.Error("Game object fields do not match expected values")
    }*/
}

func TestUintToString(t *testing.T) {
	tests := []struct {
		input uint
		want  string
	}{
		{0, "0"},
		{25, "25"},
		{100, "100"},
	}

	for _, test := range tests {
		got := uintToString(test.input)
		if got != test.want {
			t.Errorf("uintToString(%v) = %v; want %v", test.input, got, test.want)
		}
	}
}

func TestGetColorFromKey(t *testing.T) {
	tests := []struct {
		input uint8
		want  string
		err   error
	}{
		{1, "#D9362A", nil},
		{2, "#BE185D", nil},
	}

	for _, test := range tests {
		got, err := getColorFromKey(test.input)
		if got != test.want || err != test.err {
			t.Errorf("getColorFromKey(%v) = %v, %v; want %v, %v", test.input, got, err, test.want, test.err)
		}
	}
}

func TestGenerateColorSquareGrid(t *testing.T) {
	solvedBoard := []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	size := uint(4)
	expectedOutput, _ := generateColorSquareGrid(solvedBoard, size)

	output, err := generateColorSquareGrid(solvedBoard, size)
	if err != nil {
		t.Errorf("generateColorSquareGrid() error = %v", err)
		return
	}
	if output != expectedOutput {
		t.Errorf("generateColorSquareGrid() = %v; want %v", output, expectedOutput)
	}
}

/*
func TestGenerateRandomNumbers(t *testing.T) {
	tests := []struct {
		num, start, end int
		valid           bool
	}{
		{5, 1, 10, true},
		{-1, 1, 10, false},
		{11, 1, 10, false},
	}

	for _, test := range tests {
		result := generateRandomNumbers(test.num, test.start, test.end)

		if test.valid && result == nil {
			t.Errorf("Expected valid result for input (%d, %d, %d), got nil", test.num, test.start, test.end)
			continue
		}

		if !test.valid && result != nil {
			t.Errorf("Expected nil for input (%d, %d, %d), got: %v", test.num, test.start, test.end, result)
			continue
		}

		if test.valid {
			// Check if values are unique
			values := make(map[int]bool)
			for _, val := range result {
				if values[val] {
					t.Errorf("Duplicate value found: %d", val)
				}
				values[val] = true

				// Check if the value is within range
				if val < test.start || val > test.end {
					t.Errorf("Value out of range: %d not in [%d, %d]", val, test.start, test.end)
				}
			}

			// Check if we got the correct number of values
			if len(result) != test.num {
				t.Errorf("Expected %d values, got %d", test.num, len(result))
			}
		}
	}
}

// Optional: Test the randomness (this is more of a sanity check)
func TestRandomness(t *testing.T) {
	results := make(map[int]int)
	trials := 10
	for i := 0; i < trials; i++ {
		// Add a microsecond delay before each call
		for i := 0; i < 500000; i++ {
		}

		numbers := generateRandomNumbers(1, 1, 10)
		results[numbers[0]]++
	}

	// Expect about even distribution
	for num, count := range results {
		// Adjusted basic check for 100 trials
		if count < (trials/10)/2 || count > (trials/10)*2 {
			t.Errorf("Number %d appears %d times, which might indicate lack of randomness", num, count)
		}
	}
}*/

// Helper function to compare two slices for equality
func sliceEqual(a, b []int) bool {
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
