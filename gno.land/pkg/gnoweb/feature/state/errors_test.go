package state

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

// mapClientError must surface ErrClientResponseTooLarge as a clear 502,
// not bury it as a generic internal-error.
func TestMapClientErrorTooLarge(t *testing.T) {
	tooLarge := errors.New("RPC node response too large: 9000000 bytes (max 8388608)")
	status, msg := mapClientError(tooLarge)
	if status != http.StatusBadGateway {
		t.Errorf("status = %d, want 502 (BadGateway)", status)
	}
	if !strings.Contains(msg, "too large") {
		t.Errorf("message = %q, want it to mention \"too large\"", msg)
	}
}
