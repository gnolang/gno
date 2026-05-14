package state

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Sentinel error message fragments. The feature/state package cannot import
// gno.land/pkg/gnoweb (that would create an import cycle), so it cannot use
// errors.Is against gnoweb's ErrClient* sentinels. Instead, mapClientError
// matches on the stable error-message substrings produced by gnoweb's client
// (client.go:19-25). If those strings ever change, both sides must move in
// lockstep — keep this list aligned.
const (
	clientErrPackageNotFound = "package not found"
	clientErrObjectNotFound  = "object not found"
	clientErrTimeout         = "RPC node request timeout"
	clientErrBadRequest      = "bad request"
)

// mapClientError classifies a ClientAdapter error into (status, friendly msg).
// Mirrors handler_http.go:clientErrorMessage — height > 0 surfaces "block
// height N is not available" for non-NotFound errors so an out-of-range pin
// does not surface as a generic 500. NotFound wins regardless of height (a
// wrong path is wrong at any block). Timeout → 408, bad-request → 400,
// everything else → 500 with internals hidden.
func mapClientError(err error, height int64) (status int, message string) {
	if err == nil {
		return http.StatusOK, ""
	}
	// Substring match (not errors.Is) — see sentinel constants above.
	msg := err.Error()
	if strings.Contains(msg, clientErrPackageNotFound) || strings.Contains(msg, clientErrObjectNotFound) {
		return http.StatusNotFound, msg
	}
	if height > 0 {
		return http.StatusBadRequest, fmt.Sprintf("block height %d is not available", height)
	}
	switch {
	case strings.Contains(msg, clientErrTimeout):
		return http.StatusRequestTimeout, msg
	case strings.Contains(msg, clientErrBadRequest):
		return http.StatusBadRequest, "bad request"
	default:
		return http.StatusInternalServerError, "internal error"
	}
}

// writeJSONError writes the standard `{"error":"…"}` envelope at the given
// status. Consumers parse this without sniffing for HTML.
func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	body, _ := json.Marshal(map[string]string{"error": message})
	w.Write(body)
}
