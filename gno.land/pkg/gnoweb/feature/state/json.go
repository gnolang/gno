package state

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// pkgJSONWrapper is the paginated `?state&json` response shape. Names /
// Values stay as opaque sub-arrays of the upstream qpkg_json so clients
// keep their amino-aware decoders untouched. `total` lets clients iterate
// the full realm by hopping `offset`/`limit` in cache-friendly steps.
type pkgJSONWrapper struct {
	PkgPath string            `json:"pkg_path"`
	Height  int64             `json:"height"`
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
// JSON endpoints write directly to w and always return a nil view; the
// wire-in interprets a nil view as "body already written, status set".
func (h *Handler) serveJSON(ctx context.Context, w http.ResponseWriter, r *http.Request, u *weburl.GnoURL) (int, *components.View) {
	height := u.Height()
	switch {
	case u.WebQuery.Has("oid"):
		oid := u.WebQuery.Get("oid")
		if ValidateOID(oid) != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid object id")
			return http.StatusBadRequest, nil
		}
		return h.serveJSONPassthrough(ctx, w, u, height, func(c context.Context) ([]byte, error) {
			return h.deps.Client.StateObject(c, oid, height)
		})
	case u.WebQuery.Has("tid"):
		tid := u.WebQuery.Get("tid")
		if ValidateTID(tid) != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid type id")
			return http.StatusBadRequest, nil
		}
		return h.serveJSONPassthrough(ctx, w, u, height, func(c context.Context) ([]byte, error) {
			return h.deps.Client.StateType(c, tid, height)
		})
	default:
		return h.serveJSONPackage(ctx, w, u, height)
	}
}

// serveJSONPassthrough writes the upstream amino JSON as-is for the
// already-bounded oid/tid endpoints (single object / single type, capped
// by maxRPCResponseSize on the fetch side).
func (h *Handler) serveJSONPassthrough(ctx context.Context, w http.ResponseWriter, u *weburl.GnoURL, height int64, fetch func(context.Context) ([]byte, error)) (int, *components.View) {
	raw, err := fetch(ctx)
	if err != nil {
		h.deps.Logger.Error("unable to fetch state json", "error", err, "path", u.EncodeURL(), "height", height)
		status, msg := mapClientError(err, height)
		writeJSONError(w, status, msg)
		return status, nil
	}
	writeJSONHeaders(w, height)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
	return http.StatusOK, nil
}

// serveJSONPackage returns the paginated `?state&json` view of qpkg_json.
// Names/Values stay as opaque json.RawMessage so the upstream amino bytes
// pass through untouched, wrapped in a {total, offset, limit, …} envelope.
func (h *Handler) serveJSONPackage(ctx context.Context, w http.ResponseWriter, u *weburl.GnoURL, height int64) (int, *components.View) {
	offset, err := ValidateOffset(u.WebQuery.Get("offset"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid offset")
		return http.StatusBadRequest, nil
	}
	limit, err := ValidateLimit(u.WebQuery.Get("limit"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid limit")
		return http.StatusBadRequest, nil
	}

	raw, err := h.deps.Client.StatePkg(ctx, u.Path, height)
	if err != nil {
		h.deps.Logger.Error("unable to fetch state json", "error", err, "path", u.EncodeURL(), "height", height)
		status, msg := mapClientError(err, height)
		writeJSONError(w, status, msg)
		return status, nil
	}

	var parsed struct {
		Names  []string          `json:"names"`
		Values []json.RawMessage `json:"values"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		h.deps.Logger.Error("decode pkg JSON for pagination", "error", err, "path", u.EncodeURL())
		writeJSONError(w, http.StatusBadGateway, "upstream returned malformed JSON")
		return http.StatusBadGateway, nil
	}

	total := min(len(parsed.Names), len(parsed.Values))
	start, end := clampSliceWindow(offset, limit, total)

	body, err := json.Marshal(pkgJSONWrapper{
		PkgPath: u.Path,
		Height:  height,
		Total:   total,
		Offset:  start,
		Limit:   limit,
		Names:   parsed.Names[start:end],
		Values:  parsed.Values[start:end],
	})
	if err != nil {
		h.deps.Logger.Error("marshal paginated pkg JSON", "error", err, "path", u.EncodeURL())
		writeJSONError(w, http.StatusInternalServerError, "failed to encode response")
		return http.StatusInternalServerError, nil
	}

	writeJSONHeaders(w, height)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
	return http.StatusOK, nil
}

// writeJSONHeaders stamps the canonical headers for a successful JSON
// state response. Cache-Control comes from the shared cacheControlForHeight
// so pinned/latest semantics match the HTML page and fragment surfaces.
func writeJSONHeaders(w http.ResponseWriter, height int64) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	w.Header().Set("Cache-Control", cacheControlForHeight(height))
}
