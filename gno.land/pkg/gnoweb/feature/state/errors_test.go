package state

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

// mapClientError must surface ErrClientResponseTooLarge as a clear 502,
// not bury it as a generic internal-error. Every variant below carries that
// sentinel: gnoweb's own maxRPCResponseSize cap, and — since the query export
// guards landed — the node-side size and depth rejections the gnoweb client
// remaps onto the same sentinel (client.go). All must be 502, never a generic 500.
func TestMapClientErrorTooLarge(t *testing.T) {
	for _, msg := range []string{
		"RPC node response too large: 9000000 bytes (max 8388608)",         // gnoweb cap
		"RPC node response too large: rejected by node export size guard",  // node size guard
		"RPC node response too large: rejected by node export depth guard", // node depth guard
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
