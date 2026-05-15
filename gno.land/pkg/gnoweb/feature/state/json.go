package state

import (
	"context"
	"net/http"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

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

	var (
		raw []byte
		err error
	)
	switch {
	case u.WebQuery.Has("oid"):
		oid := u.WebQuery.Get("oid")
		if ValidateOID(oid) != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid object id")
			return http.StatusBadRequest, nil
		}
		raw, err = h.deps.Client.StateObject(ctx, oid, height)
	case u.WebQuery.Has("tid"):
		tid := u.WebQuery.Get("tid")
		if ValidateTID(tid) != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid type id")
			return http.StatusBadRequest, nil
		}
		raw, err = h.deps.Client.StateType(ctx, tid, height)
	default:
		raw, err = h.deps.Client.StatePkg(ctx, u.Path, height)
	}

	if err != nil {
		h.deps.Logger.Error("unable to fetch state json", "error", err, "path", u.EncodeURL(), "height", height)
		status, msg := mapClientError(err, height)
		writeJSONError(w, status, msg)
		return status, nil
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Machine-readable surface; don't index per-height snapshots as web pages.
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	// Pinned `?height=N` is immutable once the block is finalized; "latest"
	// gets a 1s freshness window matching the ~3s block time. Sets the
	// terrain for the planned nginx/ETag layer.
	if height > 0 {
		w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=1")
	}
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
	return http.StatusOK, nil
}
