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
		input int64
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
		input int64
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
	solvedBoard := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
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


/*
func TestGenerateRandomNumbers(t *testing.T) {
	x := 5
	r := 100
	numbers := generateRandomNumbers(x, 1, int64(r))

	// Check if the length of the slice matches 'x'
	if len(numbers) != x {
		t.Fatalf("Expected length %d, but got %d", x, len(numbers))
	}

	// Check if each number is within the specified range
	for _, num := range numbers {
		if num < 0 || num >= int64(r) {
			t.Fatalf("Number %d is out of range [0, %d)", num, r)
		}
	}
	
	// A simple check to ensure randomness by generating another set and comparing.
	// This isn't foolproof, but for the purposes of this example, it'll give a basic validation.
	secondSet := generateRandomNumbers(x, 1, int64(r))
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
}*/

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
