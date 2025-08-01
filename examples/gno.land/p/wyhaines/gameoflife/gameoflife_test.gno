package gameoflife

import (
	"testing"
)

// Helper function to convert string representation to grid
func stringToGrid(s string) [][]rune {
	if s == "" {
		return [][]rune{}
	}

	rows := [][]rune{}
	currentRow := []rune{}

	for _, char := range s {
		if char == '\n' {
			rows = append(rows, currentRow)
			currentRow = []rune{}
		} else {
			currentRow = append(currentRow, char)
		}
	}

	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	return rows
}

// Helper function to convert grid to string
func gridToString(grid [][]rune) string {
	if len(grid) == 0 {
		return "_EMPTY_"
	}

	var result string
	for i, row := range grid {
		result += string(row)
		if i < len(grid)-1 {
			result += "\n"
		}
	}
	return result
}

func TestGameOfLife(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		desc     string
	}{
		// Basic patterns
		{
			name:     "Empty grid",
			input:    "",
			expected: "",
			desc:     "Empty grid remains empty",
		},
		{
			name:     "Single cell dies",
			input:    "X",
			expected: " ",
			desc:     "Single cell dies from underpopulation",
		},
		{
			name:     "Block pattern stays stable",
			input:    "XX\nXX",
			expected: "XX\nXX",
			desc:     "Block pattern is stable with each cell having 3 neighbors",
		},
		{
			name:     "Blinker pattern oscillates",
			input:    "     \n     \n XXX \n     \n     ",
			expected: "     \n  X  \n  X  \n  X  \n     ",
			desc:     "Blinker pattern oscillates between horizontal and vertical",
		},
		{
			name:     "Mixed cell types",
			input:    "     \n     \n A#o \n     \n     ",
			expected: "     \n  A  \n  #  \n  A  \n     ",
			desc:     "Different cell types interact while maintaining their character",
		},
		{
			name:     "Glider pattern",
			input:    "     \n  X  \n   X \n XXX \n     ",
			expected: "     \n     \n X X \n  XX \n  X  ",
			desc:     "Glider pattern moves diagonally across the grid",
		},
		// Cell behavior cases
		{
			name:     "Underpopulation",
			input:    "X X\n   \n   ",
			expected: "   \n   \n   ",
			desc:     "Cells with fewer than 2 neighbors die",
		},
		{
			name:     "Overpopulation",
			input:    "     \n XXX \n XXX \n XXX \n     ",
			expected: "  X  \n X X \nX   X\n X X \n  X  ",
			desc:     "Cells with more than 3 neighbors die",
		},
		{
			name:     "Survival",
			input:    "XX \nXX \n   ",
			expected: "XX \nXX \n   ",
			desc:     "Cells with 2 or 3 neighbors survive",
		},
		{
			name:     "Birth",
			input:    "     \n XXX \n XXX \n XXX \n     ",
			expected: "  X  \n X X \nX   X\n X X \n  X  ",
			desc:     "Dead cells with exactly 3 neighbors become alive",
		},
		// Wraparound cases
		{
			name:     "Wrap horizontal",
			input:    "X X\n   \n   ",
			expected: "   \n   \n   ",
			desc:     "Cells wrap horizontally for neighbor counting",
		},
		{
			name:     "Wrap vertical",
			input:    "X  \n   \nX  ",
			expected: "   \n   \n   ",
			desc:     "Cells wrap vertically for neighbor counting",
		},
		{
			name:     "Wrap corner",
			input:    "X  \n   \n  X",
			expected: "   \n   \n   ",
			desc:     "Cells wrap across corners for neighbor counting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := stringToGrid(tt.input)
			expected := stringToGrid(tt.expected)
			result := NextGeneration(input)

			resultStr := gridToString(result)
			expectedStr := gridToString(expected)

			if resultStr != expectedStr {
				t.Errorf("\n%s\nexpected:\n%s\ngot:\n%s", tt.desc, expectedStr, resultStr)
			}
		})
	}
}

func TestWraparound(t *testing.T) {
	input := stringToGrid("X  \n   \n  X")
	result := NextGeneration(input)

	// The corners should interact due to wraparound
	if result[0][0] != ' ' {
		t.Error("Expected top-left corner to die due to underpopulation")
	}
	if result[2][2] != ' ' {
		t.Error("Expected bottom-right corner to die due to underpopulation")
	}
}


func TestCharacterInheritance(t *testing.T) {
	input := stringToGrid("A#@\n # \n   ")
	result := NextGeneration(input)

	// The middle cell should inherit the most common neighbor character
	if result[1][1] != '#' {
		t.Errorf("Expected middle cell to inherit '#' as the most common neighbor, got '%c'", result[1][1])
	}
}

func TestTwoByTwoGrid(t *testing.T) {
	input := stringToGrid("XX\nXX")
	result := NextGeneration(input)

	// Each cell should survive as it has exactly 3 neighbors
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			if result[i][j] == ' ' {
				t.Errorf("Cell at (%d,%d) should be alive but is dead", i, j)
			}
		}
	}
}
