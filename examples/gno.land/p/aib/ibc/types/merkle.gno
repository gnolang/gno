package types

// TODO import from ics23
type MerklePath struct {
	KeyPath [][]byte
}

// BuildMerklePath takes the merkle path prefix and an ICS24 path
// and builds a new path by appending the ICS24 path to the last element of the
// merkle path prefix.
func BuildMerklePath(prefix [][]byte, path []byte) MerklePath {
	prefixLength := len(prefix)
	if prefixLength == 0 {
		panic("cannot build merkle path with empty prefix")
	}

	// copy prefix to avoid modifying the original slice
	fullPath := make([][]byte, len(prefix))
	for i, p := range prefix {
		fullPath[i] = make([]byte, len(p))
		copy(fullPath[i], p)
	}
	// append path to last element
	fullPath[prefixLength-1] = append(fullPath[prefixLength-1], path...)
	return MerklePath{KeyPath: fullPath}
}
