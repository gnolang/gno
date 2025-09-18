package types

import "gno.land/p/demo/ufmt"

// Height is a monotonically increasing data type
// that can be compared against another Height for the purposes of updating and
// freezing clients
//
// Normally the RevisionHeight is incremented at each height while keeping
// RevisionNumber the same. However some consensus algorithms may choose to
// reset the height in certain conditions e.g. hard forks, state-machine
// breaking changes In these cases, the RevisionNumber is incremented so that
// height continues to be monitonically increasing even as the RevisionHeight
// gets reset
//
// Please note that json tags for generated Go code are overridden to explicitly exclude the omitempty jsontag.
// This enforces the Go json marshaller to always emit zero values for both revision_number and revision_height.
type Height struct {
	// the revision that the client is currently on
	RevisionNumber uint64 `protobuf:"varint,1,opt,name=revision_number,json=revisionNumber,proto3" json:"revision_number"`
	// the height within the given revision
	RevisionHeight uint64 `protobuf:"varint,2,opt,name=revision_height,json=revisionHeight,proto3" json:"revision_height"`
}

func (h Height) String() string {
	return ufmt.Sprintf("%d/%d", h.RevisionNumber, h.RevisionHeight)
}

// GTE Helper comparison function returns true if h >= other
func (h Height) GTE(other Height) bool {
	cmp := h.Compare(other)
	return cmp >= 0
}

// GT Helper comparison function returns true if h > other
func (h Height) GT(other Height) bool {
	return h.Compare(other) == 1
}

// IsZero returns true if height revision and revision-height are both 0
func (h Height) IsZero() bool {
	return h.RevisionNumber == 0 && h.RevisionHeight == 0
}

// Compare implements a method to compare two heights. When comparing two heights a, b
// we can call a.Compare(b) which will return
// -1 if a < b
// 0  if a = b
// 1  if a > b
//
// It first compares based on revision numbers, whichever has the higher revision number is the higher height
// If revision number is the same, then the revision height is compared
func (h Height) Compare(height Height) int64 {
	if h.RevisionNumber < height.RevisionNumber {
		return -1
	}
	if h.RevisionNumber > height.RevisionNumber {
		return 1
	}
	if h.RevisionHeight < height.RevisionHeight {
		return -1
	}
	if h.RevisionHeight > height.RevisionHeight {
		return 1
	}
	return 0
}
