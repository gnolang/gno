package state

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

// mapClientError must surface ErrClientResponseTooLarge as a clear 502,
// not bury it as a generic internal-error. Both variants below carry that
// sentinel: gnoweb's own maxRPCResponseSize cap, and — since the query export
// size guard landed — a node-side rejection the gnoweb client remaps onto the
// same sentinel (client.go). Both must be 502, never a generic 500.
func TestMapClientErrorTooLarge(t *testing.T) {
	for _, msg := range []string{
		"RPC node response too large: 9000000 bytes (max 8388608)",        // gnoweb cap
		"RPC node response too large: rejected by node export size guard", // node guard
	} {
		status, got := mapClientError(errors.New(msg))
		if status != http.StatusBadGateway {
			t.Errorf("status = %d, want 502 (BadGateway) for %q", status, msg)
		}
		if !strings.Contains(got, "too large") {
			t.Errorf("message = %q, want it to mention \"too large\"", got)
		}
	}
}
