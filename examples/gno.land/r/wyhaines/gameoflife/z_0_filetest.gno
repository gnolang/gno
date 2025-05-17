package main

import (
	"fmt"

	"gno.land/r/wyhaines/gameoflife"
)

func main() {
	// Test empty grid
	fmt.Printf("Empty grid:\n%s\n\n", gameoflife.Render(""))

	// Test single cell (should die)
	fmt.Printf("Single cell:\nInput:\nX\nOutput:\n%s\n\n", gameoflife.Render("X"))

	// Test stable block pattern
	input := "XX\\nXX"
	fmt.Printf("Block pattern:\nInput:\n%s\nOutput:\n%s\n\n", input, gameoflife.Render(input))

	// Test blinker pattern (oscillator)
	input = "     \\n     \\n XXX \\n     \\n     "
	fmt.Printf("Blinker pattern:\nInput:\n%s|\nOutput:\n%s|\n", input, gameoflife.Render(input))

	// Test mixed cell types
	input = "     \\n     \\n A#o \\n     \\n     "
	fmt.Printf("Mixed cells:\nInput:\n%s|\nOutput:\n%s|\n", input, gameoflife.Render(input))

	// Test glider pattern
	input = "     \\n  X  \\n   X \\n XXX \\n     "
	fmt.Printf("Glider pattern:\nInput:\n%s|\nOutput:\n%s|\n", input, gameoflife.Render(input))
}

// Output:
// Empty grid:
//
//
// Single cell:
// Input:
// X
// Output:
//
//
// Block pattern:
// Input:
// XX\nXX
// Output:
// XX\nXX
//
// Blinker pattern:
// Input:
//      \n     \n XXX \n     \n     |
// Output:
//      \n  X  \n  X  \n  X  \n     |
// Mixed cells:
// Input:
//      \n     \n A#o \n     \n     |
// Output:
//      \n  A  \n  #  \n  A  \n     |
// Glider pattern:
// Input:
//      \n  X  \n   X \n XXX \n     |
// Output:
//      \n     \n X X \n  XX \n  X  |
