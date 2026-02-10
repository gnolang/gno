// The diff package implements the Myers diff algorithm to compute the edit distance
// and generate a minimal edit script between two strings.
//
// Edit distance, also known as Levenshtein distance, is a measure of the similarity
// between two strings. It is defined as the minimum number of single-character edits (insertions,
// deletions, or substitutions) required to change one string into the other.
package diff

import (
	"strings"
)

// EditType represents the type of edit operation in a diff.
type EditType uint8

const (
	// EditKeep indicates that a character is unchanged in both strings.
	EditKeep EditType = iota

	// EditInsert indicates that a character was inserted in the new string.
	EditInsert

	// EditDelete indicates that a character was deleted from the old string.
	EditDelete
)

// Edit represent a single edit operation in a diff.
type Edit struct {
	// Type is the kind of edit operation.
	Type EditType

	// Char is the character involved in the edit operation.
	Char rune
}

// MyersDiff computes the difference between two strings using Myers' diff algorithm.
// It returns a slice of Edit operations that transform the old string into the new string.
// This implementation finds the shortest edit script (SES) that represents the minimal
// set of operations to transform one string into the other.
//
// The function handles both ASCII and non-ASCII characters correctly.
//
// Time complexity: O((N+M)D), where N and M are the lengths of the input strings,
// and D is the size of the minimum edit script.
//
// Space complexity: O((N+M)D)
//
// In the worst case, where the strings are completely different, D can be as large as N+M,
// leading to a time and space complexity of O((N+M)^2). However, for strings with many
// common substrings, the performance is much better, often closer to O(N+M).
//
// Parameters:
//   - old: the original string.
//   - new: the modified string.
//
// Returns:
//   - A slice of Edit operations representing the minimum difference between the two strings.
func MyersDiff(old, new string) []Edit {
	oldRunes, newRunes := []rune(old), []rune(new)
	n, m := len(oldRunes), len(newRunes)

	if n == 0 && m == 0 {
		return []Edit{}
	}

	// old is empty
	if n == 0 {
		edits := make([]Edit, m)
		for i, r := range newRunes {
			edits[i] = Edit{Type: EditInsert, Char: r}
		}
		return edits
	}

	if m == 0 {
		edits := make([]Edit, n)
		for i, r := range oldRunes {
			edits[i] = Edit{Type: EditDelete, Char: r}
		}
		return edits
	}

	max := n + m
	v := make([]int, 2*max+1)
	var trace [][]int
search:
	for d := 0; d <= max; d++ {
		// iterate through diagonals
		for k := -d; k <= d; k += 2 {
			var x int
			if k == -d || (k != d && v[max+k-1] < v[max+k+1]) {
				x = v[max+k+1] // move down
			} else {
				x = v[max+k-1] + 1 // move right
			}
			y := x - k

			// extend the path as far as possible with matching characters
			for x < n && y < m && oldRunes[x] == newRunes[y] {
				x++
				y++
			}

			v[max+k] = x

			// check if we've reached the end of both strings
			if x == n && y == m {
				trace = append(trace, append([]int(nil), v...))
				break search
			}
		}
		trace = append(trace, append([]int(nil), v...))
	}

	// backtrack to construct the edit script
	edits := make([]Edit, 0, n+m)
	x, y := n, m
	for d := len(trace) - 1; d >= 0; d-- {
		vPrev := trace[d]
		k := x - y
		var prevK int
		if k == -d || (k != d && vPrev[max+k-1] < vPrev[max+k+1]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}
		prevX := vPrev[max+prevK]
		prevY := prevX - prevK

		// add keep edits for matching characters
		for x > prevX && y > prevY {
			if x > 0 && y > 0 {
				edits = append([]Edit{{Type: EditKeep, Char: oldRunes[x-1]}}, edits...)
			}
			x--
			y--
		}
		if y > prevY {
			if y > 0 {
				edits = append([]Edit{{Type: EditInsert, Char: newRunes[y-1]}}, edits...)
			}
			y--
		} else if x > prevX {
			if x > 0 {
				edits = append([]Edit{{Type: EditDelete, Char: oldRunes[x-1]}}, edits...)
			}
			x--
		}
	}

	return edits
}

// Format converts a slice of Edit operations into a human-readable string representation.
// It groups consecutive edits of the same type and formats them as follows:
//   - Unchanged characters are left as-is
//   - Inserted characters are wrapped in [+...]
//   - Deleted characters are wrapped in [-...]
//
// This function is useful for visualizing the differences between two strings
// in a compact and intuitive format.
//
// Parameters:
//   - edits: A slice of Edit operations, typically produced by MyersDiff
//
// Returns:
//   - A formatted string representing the diff
//
// Example output:
//
//	For the diff between "abcd" and "acbd", the output might be:
//	"a[-b]c[+b]d"
//
// Note:
//
//	The function assumes that the input slice of edits is in the correct order.
//	An empty input slice will result in an empty string.
func Format(edits []Edit) string {
	if len(edits) == 0 {
		return ""
	}

	var (
		result       strings.Builder
		currentType  EditType
		currentChars strings.Builder
	)

	flushCurrent := func() {
		if currentChars.Len() > 0 {
			switch currentType {
			case EditKeep:
				result.WriteString(currentChars.String())
			case EditInsert:
				result.WriteString("[+")
				result.WriteString(currentChars.String())
				result.WriteByte(']')
			case EditDelete:
				result.WriteString("[-")
				result.WriteString(currentChars.String())
				result.WriteByte(']')
			}
			currentChars.Reset()
		}
	}

	for _, edit := range edits {
		if edit.Type != currentType {
			flushCurrent()
			currentType = edit.Type
		}
		currentChars.WriteRune(edit.Char)
	}
	flushCurrent()

	return result.String()
}
