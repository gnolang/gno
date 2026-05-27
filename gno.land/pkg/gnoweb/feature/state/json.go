package state

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// pkgJSONWrapper is the paginated `?state&json` response shape. Names /
// Values stay as opaque sub-arrays of the upstream qpkg_json so clients
// keep their amino-aware decoders untouched. `total` lets clients iterate
// the full realm by hopping `offset`/`limit` in cache-friendly steps.
type pkgJSONWrapper struct {
	PkgPath string            `json:"pkg_path"`
	Total   int               `json:"total"`
	Offset  int               `json:"offset"`
	Limit   int               `json:"limit"`
	Names   []string          `json:"names"`
	Values  []json.RawMessage `json:"values"`
}

// serveJSON exposes the chain's raw Amino JSON as a stable API surface for
// external tooling (block explorers, IDE plugins, JS SDKs) that decode state
// in the browser. Triggered by `?state&json`, `?state&oid=…&json`, or
// `?state&tid=…&json`.
//
// Bytes flow through unmodified: no decoder, no walker, no fan-out, so none
// of the per-render bounds apply. Validation happens before any RPC so a
// malformed oid/tid never reaches the chain.
//
// JSON endpoints write directly to w; the caller wraps the int status
// with a nil view since the wire-in interprets nil as "body already
// written, status set".
func (h *Handler) serveJSON(ctx context.Context, w http.ResponseWriter, u *weburl.GnoURL) int {
	switch {
	case u.WebQuery.Has("oid"):
		oid := u.WebQuery.Get("oid")
		if ValidateOID(oid) != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid object id")
			return http.StatusBadRequest
		}
		return h.serveJSONPassthrough(ctx, w, u, func(c context.Context) ([]byte, error) {
			return h.deps.Client.StateObject(c, oid, 0)
		})
	case u.WebQuery.Has("tid"):
		tid := u.WebQuery.Get("tid")
		if ValidateTID(tid) != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid type id")
			return http.StatusBadRequest
		}
		return h.serveJSONPassthrough(ctx, w, u, func(c context.Context) ([]byte, error) {
			return h.deps.Client.StateType(c, tid, 0)
		})
	default:
		return h.serveJSONPackage(ctx, w, u)
	}
}

// serveJSONPassthrough writes the upstream amino JSON as-is for the
// already-bounded oid/tid endpoints (single object / single type, capped
// by maxRPCResponseSize on the fetch side).
func (h *Handler) serveJSONPassthrough(ctx context.Context, w http.ResponseWriter, u *weburl.GnoURL, fetch func(context.Context) ([]byte, error)) int {
	raw, err := fetch(ctx)
	if err != nil {
		h.deps.Logger.Error("unable to fetch state json", "error", err, "path", u.EncodeURL())
		status, msg := mapClientError(err)
		writeJSONError(w, status, msg)
		return status
	}
	writeJSONHeaders(w)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
	return http.StatusOK
}

// serveJSONPackage returns the paginated `?state&json` view of qpkg_json.
// Names/Values stay as opaque json.RawMessage so the upstream amino bytes
// pass through untouched, wrapped in a {total, offset, limit, …} envelope.
func (h *Handler) serveJSONPackage(ctx context.Context, w http.ResponseWriter, u *weburl.GnoURL) int {
	offset, err := ValidateOffset(u.WebQuery.Get("offset"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid offset")
		return http.StatusBadRequest
	}
	limit, err := ValidateLimit(u.WebQuery.Get("limit"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid limit")
		return http.StatusBadRequest
	}

	raw, err := h.deps.Client.StatePkg(ctx, u.Path, 0)
	if err != nil {
		h.deps.Logger.Error("unable to fetch state json", "error", err, "path", u.EncodeURL())
		status, msg := mapClientError(err)
		writeJSONError(w, status, msg)
		return status
	}

	var parsed struct {
		Names  []string          `json:"names"`
		Values []json.RawMessage `json:"values"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		h.deps.Logger.Error("decode pkg JSON for pagination", "error", err, "path", u.EncodeURL())
		writeJSONError(w, http.StatusBadGateway, "upstream returned malformed JSON")
		return http.StatusBadGateway
	}

	total := min(len(parsed.Names), len(parsed.Values))
	start, end := clampSliceWindow(offset, limit, total)

	body, err := json.Marshal(pkgJSONWrapper{
		PkgPath: u.Path,
		Total:   total,
		Offset:  start,
		Limit:   limit,
		Names:   parsed.Names[start:end],
		Values:  parsed.Values[start:end],
	})
	if err != nil {
		h.deps.Logger.Error("marshal paginated pkg JSON", "error", err, "path", u.EncodeURL())
		writeJSONError(w, http.StatusInternalServerError, "failed to encode response")
		return http.StatusInternalServerError
	}

	writeJSONHeaders(w)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
	return http.StatusOK
}

// writeJSONHeaders stamps the canonical headers for a successful JSON
// state response. Mirrors the HTML page and fragment surfaces.
func writeJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	w.Header().Set("Cache-Control", stateCacheControl)
	w.Header().Set("Vary", "HX-Request")
}
