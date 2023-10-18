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

func TestintToString(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{25, "25"},
		{100, "100"},
	}

	for _, test := range tests {
		got := intToString(test.input)
		if got != test.want {
			t.Errorf("intToString(%v) = %v; want %v", test.input, got, test.want)
		}
	}
}

func TestGetColorFromKey(t *testing.T) {
	tests := []struct {
		input int
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
	solvedBoard := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	size := int(4)
	expectedOutput, _ := GenerateColorSquareGrid(solvedBoard, size)

	output, err := GenerateColorSquareGrid(solvedBoard, size)
	if err != nil {
		t.Errorf("GenerateColorSquareGrid() error = %v", err)
		return
	}
	if output != expectedOutput {
		t.Errorf("GenerateColorSquareGrid() = %v; want %v", output, expectedOutput)
	}
}

func TestGenerateRandomNumbers(t *testing.T) {
	x := 5
	r := 100
	numbers := generateRandomNumbers(x, r)

	// Check if the length of the slice matches 'x'
	if len(numbers) != x {
		t.Fatalf("Expected length %d, but got %d", x, len(numbers))
	}

	// Check if each number is within the specified range
	for _, num := range numbers {
		if num < 0 || num >= r {
			t.Fatalf("Number %d is out of range [0, %d)", num, r)
		}
	}

	// A simple check to ensure randomness by generating another set and comparing.
	// This isn't foolproof, but for the purposes of this example, it'll give a basic validation.
	secondSet := generateRandomNumbers(x, r)
	sameCount := 0
	for i := 0; i < x; i++ {
		if numbers[i] == secondSet[i] {
			sameCount++
		}
	}

	// If all the numbers are the same in a subsequent set, that's suspicious!
	if sameCount == x {
		t.Fatal("Two subsequent sets of random numbers are the same. Randomness might be compromised.")
	}
}

/*
func TestParseString(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
		err      error
	}{
		{"[2,3,5,3,3,5,6,7,2,4,8,6,5,4,7]", []int{2, 3, 5, 3, 3, 5, 6, 7, 2, 4, 8, 6, 5, 4, 7}, nil},
		{"[1,2,3]", []int{1, 2, 3}, nil},
		{"[a,b,c]", nil, nil},
	}

	for _, test := range tests {
		output, err := ParseString(test.input)
		if !reflect.DeepEqual(output, test.expected) || !reflect.DeepEqual(err, test.err) {
			t.Errorf("For input '%s', expected %v, got %v with error %v", test.input, test.expected, output, err)
		}
	}
}


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
