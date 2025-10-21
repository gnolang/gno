package types

import (
	"chain/runtime"
	"regexp"
	"strconv"
	"strings"

	"gno.land/p/nt/ufmt"
)

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
	RevisionNumber uint64
	// the height within the given revision
	RevisionHeight uint64
}

func NewHeight(revision, height uint64) Height {
	return Height{
		RevisionNumber: revision,
		RevisionHeight: height,
	}
}

// ParseHeight is a utility function that takes a string representation of the
// height and returns a Height struct.
func ParseHeight(heightStr string) (Height, error) {
	splitStr := strings.Split(heightStr, "/")
	if len(splitStr) != 2 {
		return Height{}, ufmt.Errorf("expected height string format: {revision}/{height}. Got: %s", heightStr)
	}
	revisionNumber, err := strconv.ParseUint(splitStr[0], 10, 64)
	if err != nil {
		return Height{}, ufmt.Errorf("invalid revision number. parse err: %s", err)
	}
	revisionHeight, err := strconv.ParseUint(splitStr[1], 10, 64)
	if err != nil {
		return Height{}, ufmt.Errorf("invalid revision height. parse err: %s", err)
	}
	return NewHeight(revisionNumber, revisionHeight), nil
}

func ZeroHeight() Height { return Height{} }

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

// LTE Helper comparison function returns true if h <= other
func (h Height) LTE(other Height) bool {
	cmp := h.Compare(other)
	return cmp <= 0
}

// LT Helper comparison function returns true if h < other
func (h Height) LT(other Height) bool {
	return h.Compare(other) == -1
}

// EQ Helper comparison function returns true if h == other
func (h Height) EQ(other Height) bool {
	return h.Compare(other) == 0
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

// GetSelfHeight is a utility function that returns self height given context.
// Revision number is retrieved from runtime.ChainID()
func GetSelfHeight() Height {
	revision := ParseChainID(runtime.ChainID())
	return NewHeight(revision, uint64(runtime.ChainHeight()))
}

// IsRevisionFormat checks if a chainID is in the format required for parsing
// revisions
// The chainID must be in the form: `{chainID}-{revision}`.
// 24-host may enforce stricter checks on chainID
var IsRevisionFormat = regexp.MustCompile(`^.*[^\n-]-{1}[1-9][0-9]*$`).MatchString

// ParseChainID is a utility function that returns a revision number from the
// given ChainID.
// ParseChainID attempts to parse a chain id in the format:
// `{chainID}-{revision}` and return the revisionnumber as a uint64.
// If the chainID is not in the expected format, a default revision value of 0
// is returned.
func ParseChainID(chainID string) uint64 {
	if !IsRevisionFormat(chainID) {
		// chainID is not in revision format, return 0 as default
		return 0
	}
	splitStr := strings.Split(chainID, "-")
	revision, err := strconv.ParseUint(splitStr[len(splitStr)-1], 10, 64)
	// sanity check: error should always be nil since regex only allows numbers in last element
	if err != nil {
		panic(ufmt.Errorf("regex allowed non-number value as last split element for chainID: %s", chainID))
	}
	return revision
}
