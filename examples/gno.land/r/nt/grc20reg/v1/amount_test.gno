package grc20reg

import (
	"strconv"
	"testing"
)

// equal256 and require256 compare 256-bit amounts natively; uassert and
// urequire do not support uint256 (see p/nt/ufmt's README).
func equal256(t *testing.T, want, got uint256, msgs ...string) bool {
	t.Helper()
	if want == got {
		return true
	}
	t.Errorf("%swant %s, got %s", msgPrefix(msgs), strconv.FormatUint256(want, 10), strconv.FormatUint256(got, 10))
	return false
}

func require256(t *testing.T, want, got uint256, msgs ...string) {
	t.Helper()
	if want != got {
		t.Fatalf("%swant %s, got %s", msgPrefix(msgs), strconv.FormatUint256(want, 10), strconv.FormatUint256(got, 10))
	}
}

func msgPrefix(msgs []string) string {
	if len(msgs) == 0 {
		return ""
	}
	return msgs[0] + ": "
}
