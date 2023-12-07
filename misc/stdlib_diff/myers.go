package main

import (
	"slices"
)

var _ Algorithm = (*Myers)(nil)

// Myers is a struct representing the Myers algorithm for line-based difference.
type Myers struct {
	src []string // Lines of the source file.
	dst []string // Lines of the destination file.
}

// NewMyers creates a new Myers instance with the specified source and destination lines.
func NewMyers(src, dst []string) *Myers {
	return &Myers{
		src: src,
		dst: dst,
	}
}

// Do performs the Myers algorithm to find the differences between source and destination files.
// It returns the differences as two slices of LineDifferrence representing source and destination changes.
func (m *Myers) Do() ([]LineDifferrence, []LineDifferrence) {
	operations := m.doMyers()

	srcIndex, dstIndex := 0, 0
	dstDiff := make([]LineDifferrence, 0)
	srcDiff := make([]LineDifferrence, 0)
	insertCount := 0
	for _, op := range operations {
		switch op {
		case INSERT:
			dstDiff = append(dstDiff, LineDifferrence{Line: "+" + m.dst[dstIndex], Operation: op.String()})
			srcDiff = append(srcDiff, LineDifferrence{Line: "", Operation: MOVE.String()})
			dstIndex += 1
			insertCount++
			continue

		case MOVE:
			dstDiff = append(dstDiff, LineDifferrence{Line: m.src[srcIndex], Operation: op.String()})
			srcDiff = append(srcDiff, LineDifferrence{Line: m.src[srcIndex], Operation: op.String()})
			srcIndex += 1
			dstIndex += 1
			continue

		case DELETE:
			dstDiff = append(dstDiff, LineDifferrence{Line: "", Operation: MOVE.String()})
			srcDiff = append(srcDiff, LineDifferrence{Line: "-" + m.src[srcIndex], Operation: op.String()})
			srcIndex += 1
			continue
		}
	}

	// Means that src file is empty.
	if insertCount == len(srcDiff) {
		srcDiff = make([]LineDifferrence, 0)
	}
	return srcDiff, dstDiff
}

// doMyers performs the Myers algorithm and returns the list of operations.
func (m *Myers) doMyers() []operation {
	var tree []map[int]int
	var x, y int

	srcLen := len(m.src)
	dstLen := len(m.dst)
	max := srcLen + dstLen

	for pathLen := 0; pathLen <= max; pathLen++ {
		optimalCoordinates := make(map[int]int, pathLen+2)
		tree = append(tree, optimalCoordinates)

		if pathLen == 0 {
			commonPrefixLen := 0
			for srcLen > commonPrefixLen && dstLen > commonPrefixLen && m.src[commonPrefixLen] == m.dst[commonPrefixLen] {
				commonPrefixLen++
			}
			optimalCoordinates[0] = commonPrefixLen

			if commonPrefixLen == srcLen && commonPrefixLen == dstLen {
				return m.getAllOperations(tree)
			}
			continue
		}

		lastV := tree[pathLen-1]

		for k := -pathLen; k <= pathLen; k += 2 {
			if k == -pathLen || (k != pathLen && lastV[k-1] < lastV[k+1]) {
				x = lastV[k+1]
			} else {
				x = lastV[k-1] + 1
			}

			y = x - k

			for x < srcLen && y < dstLen && m.src[x] == m.dst[y] {
				x, y = x+1, y+1
			}

			optimalCoordinates[k] = x

			if x == srcLen && y == dstLen {
				return m.getAllOperations(tree)
			}
		}
	}

	return m.getAllOperations(tree)
}

// getAllOperations retrieves the list of operations from the calculated tree.
func (m *Myers) getAllOperations(tree []map[int]int) []operation {
	var operations []operation
	var k, prevK, prevX, prevY int

	x := len(m.src)
	y := len(m.dst)

	for pathLen := len(tree) - 1; pathLen > 0; pathLen-- {
		k = x - y
		lastV := tree[pathLen-1]

		if k == -pathLen || (k != pathLen && lastV[k-1] < lastV[k+1]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}

		prevX = lastV[prevK]
		prevY = prevX - prevK

		for x > prevX && y > prevY {
			operations = append(operations, MOVE)
			x -= 1
			y -= 1
		}

		if x == prevX {
			operations = append(operations, INSERT)
		} else {
			operations = append(operations, DELETE)
		}

		x, y = prevX, prevY
	}

	if tree[0][0] != 0 {
		for i := 0; i < tree[0][0]; i++ {
			operations = append(operations, MOVE)
		}
	}

	slices.Reverse(operations)
	return operations
}
