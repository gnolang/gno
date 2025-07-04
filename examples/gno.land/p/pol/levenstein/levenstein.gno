// from www.youtube.com/watch?v=d-Eq6x1yssU
// Levenstein distance is an algorithm to detect how close 2 words are.
// For example, "Float" and "Boats" have a levenstein distance of 3.
// Why ? Because you need 3 steps to go from Boats to Float
// - We start with the word (Boats)
// - Remove the s (Boat)
// - Insert the l (Bloat)
// - Substitute b with f (Float)
package levenstein

func min3(a, b, c uint) uint {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func Distance(a, b string) uint {
	// create the matrix
	rows := len(a) + 1
	cols := len(b) + 1
	matrix := make([][]uint, rows)
	for x := 0; x < rows; x++ {
		matrix[x] = make([]uint, cols)
		for y := 0; y < cols; y++ {
			// check for edges
			if x == 0 {
				matrix[x][y] = uint(y)
				continue
			} else if y == 0 {
				matrix[x][y] = uint(x)
				continue
			}

			// check for content
			if a[x-1] == b[y-1] {
				matrix[x][y] = matrix[x-1][y-1] // no cost if characters match
			} else {
				matrix[x][y] = min3(
					matrix[x][y-1]+1,   // insertion
					matrix[x-1][y]+1,   // deletion
					matrix[x-1][y-1]+1, // substitution
				)
			}
		}
	}

	for _, row := range matrix {
		println(row)
	}
	return 0
}
