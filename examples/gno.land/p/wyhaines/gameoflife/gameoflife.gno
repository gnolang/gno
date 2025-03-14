package gameoflife

// NextGeneration takes a grid (slice of slices of runes) representing the current generation
// and returns a new grid representing the next generation according to Conway's Game of Life rules.
// Rules:
// - A live cell (non-' ') with 2 or 3 live neighbors survives, otherwise dies.
// - A dead cell (' ') with exactly 3 live neighbors becomes alive, taking on the most common neighbor character.
// The grid is considered toroidal (edges wrap around). Note that when wrapping, duplicate counts
// of the same neighbor must be avoided.
func NextGeneration(grid [][]rune) [][]rune {
	// If grid is empty, return it as is.
	if len(grid) == 0 {
		return grid
	}
	height := len(grid)
	width := len(grid[0])

	// Create a new grid of the same dimensions, initializing all cells to dead (' ').
	newGrid := make([][]rune, height)
	for i := range newGrid {
		newGrid[i] = make([]rune, width)
		for j := range newGrid[i] {
			newGrid[i][j] = ' '
		}
	}

	// Process each cell in the original grid.
	for i, row := range grid {
		for j, cell := range row {
			// Build a set of unique neighbor coordinates.
			neighbors := make(map[[2]int]bool)
			for dr := -1; dr <= 1; dr++ {
				for dc := -1; dc <= 1; dc++ {
					if dr == 0 && dc == 0 {
						continue // Skip the cell itself.
					}
					ni := (i + dr + height) % height
					nj := (j + dc + width) % width
					neighbors[[2]int{ni, nj}] = true
				}
			}

			// Count live neighbors and record the frequency of each live neighbor character.
			liveCount := 0
			freq := make(map[rune]int)
			for coord := range neighbors {
				neighbor := grid[coord[0]][coord[1]]
				if neighbor != ' ' {
					liveCount++
					freq[neighbor]++
				}
			}

			// Apply Conway's Game of Life rules.
			if cell != ' ' {
				// For a live cell: it survives with 2 or 3 live neighbors.
				if liveCount == 2 || liveCount == 3 {
					newGrid[i][j] = cell
				} else {
					newGrid[i][j] = ' ' // Cell dies.
				}
			} else {
				// For a dead cell: it becomes alive if exactly 3 live neighbors exist.
				if liveCount == 3 {
					// Determine the most common neighbor character.
					var common rune
					maxCount := 0
					for r, count := range freq {
						if count > maxCount {
							maxCount = count
							common = r
						}
					}
					newGrid[i][j] = common
				} else {
					newGrid[i][j] = ' '
				}
			}
		}
	}

	return newGrid
}
