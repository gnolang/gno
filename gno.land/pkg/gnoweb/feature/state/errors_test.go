package state

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

// mapClientError must surface ErrClientResponseTooLarge as a clear 502 at
// any height, not as the misleading "block height N is not available"
// message the height>0 branch would otherwise return.
func TestMapClientErrorTooLargeBeatsHeightBranch(t *testing.T) {
	tooLarge := errors.New("RPC node response too large: 9000000 bytes (max 8388608)")

	for _, tc := range []struct {
		name   string
		height int64
	}{
		{"latest", 0},
		{"pinned", 42},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, msg := mapClientError(tooLarge, tc.height)
			if status != http.StatusBadGateway {
				t.Errorf("status = %d, want 502 (BadGateway)", status)
			}
			if !strings.Contains(msg, "too large") {
				t.Errorf("message = %q, want it to mention \"too large\"", msg)
			}
		})
	}
}
