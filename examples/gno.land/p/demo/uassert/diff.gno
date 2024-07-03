package uassert

import (
	"strings"
)

type EditType int

const (
	EditKeep EditType = iota
	EditInsert
	EditDelete
)

type Edit struct {
	Type EditType
	Char rune
}

func myersDiff(old, new string) []Edit {
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

	for d := 0; d <= max; d++ {
		for k := -d; k <= d; k += 2 {
			var x int
			if k == -d || (k != d && v[max+k-1] < v[max+k+1]) {
				x = v[max+k+1]
			} else {
				x = v[max+k-1] + 1
			}
			y := x - k

			for x < n && y < m && oldRunes[x] == newRunes[y] {
				x++
				y++
			}

			v[max+k] = x

			if x == n && y == m {
				trace = append(trace, append([]int(nil), v...))
				goto endSearch
			}
		}
		trace = append(trace, append([]int(nil), v...))
	}
endSearch:

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

func formatDiff(edits []Edit) string {
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
