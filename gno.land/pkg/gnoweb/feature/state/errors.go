package state

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Sentinel error message fragments. The feature/state package cannot import
// gno.land/pkg/gnoweb (that would create an import cycle), so it cannot use
// errors.Is against gnoweb's ErrClient* sentinels. Instead, mapClientError
// matches on the stable error-message substrings produced by gnoweb's client.
// Exported so gnoweb pins the pact in a test (see TestStateErrorSentinelPact).
const (
	ClientErrPackageNotFound  = "package not found"
	ClientErrObjectNotFound   = "object not found"
	ClientErrTimeout          = "RPC node request timeout"
	ClientErrBadRequest       = "bad request"
	ClientErrResponseTooLarge = "RPC node response too large"
)

// mapClientError classifies a ClientAdapter error into (status, friendly msg).
// Mirrors handler_http.go:clientErrorMessage. NotFound wins; Timeout → 408,
// bad-request → 400, oversize upstream → 502, everything else → 500 with
// internals hidden.
func mapClientError(err error) (status int, message string) {
	if err == nil {
		return http.StatusOK, ""
	}
	// Substring match (not errors.Is) — see sentinel constants above.
	msg := err.Error()
	if strings.Contains(msg, ClientErrPackageNotFound) || strings.Contains(msg, ClientErrObjectNotFound) {
		return http.StatusNotFound, msg
	}
	if strings.Contains(msg, ClientErrResponseTooLarge) {
		return http.StatusBadGateway, "upstream response too large"
	}
	switch {
	case strings.Contains(msg, ClientErrTimeout):
		return http.StatusRequestTimeout, msg
	case strings.Contains(msg, ClientErrBadRequest):
		return http.StatusBadRequest, "bad request"
	default:
		return http.StatusInternalServerError, "internal error"
	}
}

// writeJSONError writes the standard `{"error":"…"}` envelope at the given
// status. Consumers parse this without sniffing for HTML.
func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	body, _ := json.Marshal(map[string]string{"error": message})
	_, _ = w.Write(body)
}
