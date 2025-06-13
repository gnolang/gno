// realm.gno
package gameoflife

import (
	"strings"

	"gno.land/p/wyhaines/gameoflife"
)

// Render takes a string representing the current generation of the board,
// where each line (separated by `\n`) is one row. Each row is a slice of characters,
// with a space (' ') indicating an empty cell and any other character indicating a live cell.
// The grid's dimensions are defined by the longest line and the total number of lines.
// It parses the string into a 2D grid, computes the next generation by calling the `NextGeneration`
// function, and then returns the updated grid as a newline-separated string.
func Render(input string) string {
	// Split the input string on literal "\n" (not newline character)
	rows := strings.Split(input, "\\n")

	// Convert to [][]rune format
	grid := make([][]rune, len(rows))
	for i, row := range rows {
		grid[i] = []rune(row)
	}

	// Get next generation
	nextGen := gameoflife.NextGeneration(grid)

	// Convert back to string with newlines
	var result strings.Builder
	for i, row := range nextGen {
		result.WriteString(string(row))
		if i < len(nextGen)-1 {
			result.WriteString("\\n")
		}
	}

	return result.String()
}
