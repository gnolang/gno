package state

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// fragmentTimeout is the per-fragment deadline per ADR-004 §Resource bounds.
// Derived from the request ctx so an upstream deadline (HTTP server timeout)
// still wins; this is the upper bound a fragment can spend, not the floor.
const fragmentTimeout = 2 * time.Second

// fragSourceContextLines is the symmetrical window of lines rendered around
// the target line in frag=source. Keeps the response small while giving the
// auditor enough surrounding context to read the snippet without opening
// the full source page.
const fragSourceContextLines = 10

// maxFragmentDepth caps the attacker-controlled `depth` query param. It
// only drives a presentational --depth step-in, so a small bound is
// plenty — anything larger is absurd visual indent, not a real tree.
const maxFragmentDepth = 10

// serveFragment dispatches the fragment endpoints under one per-request
// 2 s timeout derived from the request context. Unknown `frag` values
// surface as fragment-errors (HTTP 200 + error body) so htmx swaps a
// visible message instead of silently dropping the response — see
// ADR-004 §Decision §2 fragment-error pattern.
func (h *Handler) serveFragment(ctx context.Context, w http.ResponseWriter, _ *http.Request, u *weburl.GnoURL) (int, *components.View) {
	fragCtx, cancel := context.WithTimeout(ctx, fragmentTimeout)
	defer cancel()

	switch u.WebQuery.Get("frag") {
	case "node":
		return h.serveFragNode(fragCtx, w, u)
	case "source":
		return h.serveFragSource(fragCtx, w, u)
	default:
		return writeFragError(w, "Unknown fragment type", "Please retry from the page.")
	}
}

// serveFragNode renders the immediate children of one expanded node as an
// htmx fragment. One StateObject RPC + up to fragPreviewCap nested preview
// RPCs total — the amplification budget for a single user expansion.
func (h *Handler) serveFragNode(ctx context.Context, w http.ResponseWriter, u *weburl.GnoURL) (int, *components.View) {
	oid := u.WebQuery.Get("oid")
	if err := ValidateOID(oid); err != nil {
		return writeFragError(w, "Invalid object ID", "Please retry from the page.")
	}
	height, err := ValidateHeight(u.WebQuery.Get("height"))
	if err != nil {
		return writeFragError(w, "Invalid height", "Please retry from the page.")
	}

	raw, err := h.deps.Client.StateObject(ctx, oid, height)
	if err != nil {
		return h.fragErrorFromClient(w, err, height, "oid", oid)
	}

	// Try the typed path first when the URL carries &tid= so struct
	// fields render with named members; fall back to positional on miss
	// or fetch failure. tid is opaque to gnoweb but validated against
	// the same length cap as oid to keep amplification bounded.
	tid := u.WebQuery.Get("tid")
	if tid != "" {
		if err := ValidateTID(tid); err != nil {
			tid = "" // bad tid → silent positional fallback
		}
	}
	var (
		root  StateNode
		typed bool
	)
	if tid != "" {
		if rawType, terr := h.deps.Client.StateType(ctx, tid, height); terr == nil && len(rawType) > 0 {
			if decoded, derr := DecodeObjectFull(raw, rawType, DefaultFragmentRenderConfig()); derr == nil {
				root = StateNode{Name: "(object)", Kind: KindStruct, ObjectID: oid, Children: decoded.Nodes}
				root.Length = intPtr(len(decoded.Nodes))
				typed = true
			} else {
				h.deps.Logger.Debug("frag=node typed decode failed; falling back", "oid", oid, "tid", tid, "err", derr)
			}
		}
	}
	if !typed {
		// Depth ≤3 per ADR-004 §Resource bounds; deeper exploration requires
		// the user to expand a nested node (new fragment GET).
		root, err = DecodeObject(ctx, raw, DefaultFragmentRenderConfig())
		if err != nil {
			h.deps.Logger.Error("frag=node decode failed", "oid", oid, "err", err)
			return writeFragError(w, "Could not decode object", "Please retry.")
		}
	}

	// Func/closure object: the decoded payload is a single func node
	// carrying its Source span (+ captures for closures). Promote it to
	// the root and fetch+highlight the body — one extra qfile, bounded
	// and rate-limited like any fragment — so the expansion shows the
	// actual function instead of a bare "(function): func()" row.
	if len(root.Children) == 1 {
		if c := root.Children[0]; isFuncKind(&c) {
			root = c
			if root.Source != nil && root.Source.File != "" &&
				h.deps.FileFetcher != nil && h.deps.Highlighter != nil {
				// Cap like serveFragSource: oversize files skip highlighting
				// and degrade to the lazy <details>/permalink fallback.
				if content, ferr := h.deps.FileFetcher.Fetch(ctx, u.Path, root.Source.File); ferr == nil && len(content) <= MaxFragmentFileSize {
					slice := sliceLines(content, root.Source.StartLine, root.Source.EndLine)
					if html, herr := h.deps.Highlighter.Render(root.Source.File, slice); herr == nil {
						root.SourceHTML = html
					}
				}
			}
		}
	}

	// No eager preview fetch here: ref children stay bare (ShapeRef →
	// b-state-lazy + hx-get) so the tree stays recursively drillable —
	// one StateObject RPC per click, no fan-out.
	hp := heightParam(height)
	viewMode := CanonicalViewMode(u.WebQuery.Get("view"))
	EnrichLinks(root.Children, u.Path, hp, viewMode)

	// depth is presentational only — the parent row's tree depth, so the
	// fragment's children render at depth+1 and step in via --depth.
	// Clamped to maxFragmentDepth; bad input degrades to 0 (flush-left).
	childDepth := 0
	if d, derr := strconv.Atoi(u.WebQuery.Get("depth")); derr == nil && d >= 0 {
		if d > maxFragmentDepth {
			d = maxFragmentDepth
		}
		childDepth = d + 1
	}

	writeFragSuccessHeaders(w, height)
	w.WriteHeader(http.StatusOK)
	if err := FragNodeTemplate.ExecuteTemplate(w, "fragNode", FragNodeData{
		Node:        root,
		PkgPath:     u.Path,
		Height:      height,
		HeightParam: hp,
		ViewMode:    viewMode,
		Depth:       childDepth,
	}); err != nil {
		h.deps.Logger.Error("frag=node template execute failed", "oid", oid, "err", err)
	}
	return http.StatusOK, nil
}

// serveFragSource renders a chroma'd window around the target line in the
// requested file. Capped at MaxFragmentFileSize (256 KB) — oversize files
// degrade to a fallback link to the full ?source view.
func (h *Handler) serveFragSource(ctx context.Context, w http.ResponseWriter, u *weburl.GnoURL) (int, *components.View) {
	file := u.WebQuery.Get("file")
	if err := ValidateFile(file); err != nil {
		return writeFragError(w, "Invalid file name", "")
	}
	line, err := ValidateLine(u.WebQuery.Get("line"))
	if err != nil {
		return writeFragError(w, "Invalid line number", "")
	}
	// end is optional — set by the page template for func/method spans so
	// the slice exactly matches StartLine..EndLine. Missing/invalid keeps
	// the legacy ±fragSourceContextLines window.
	var endLine int
	if endRaw := u.WebQuery.Get("end"); endRaw != "" {
		if n, verr := ValidateLine(endRaw); verr == nil && n >= line {
			endLine = n
		}
	}
	height, err := ValidateHeight(u.WebQuery.Get("height"))
	if err != nil {
		return writeFragError(w, "Invalid height", "Please retry from the page.")
	}

	if h.deps.FileFetcher == nil {
		return writeFragError(w, "Source view unavailable", "Open the file from the source tab.")
	}

	content, err := h.deps.FileFetcher.Fetch(ctx, u.Path, file)
	if err != nil {
		return h.fragErrorFromClient(w, err, height, "file", file)
	}

	heightParam := heightParam(height)

	// Oversize files fall back to a link-only message — never inline the
	// content. The template renders the b-state-frag-source skeleton with
	// an empty source body and the "See in code" permalink.
	if len(content) > MaxFragmentFileSize {
		writeFragSuccessHeaders(w, height)
		w.WriteHeader(http.StatusOK)
		_ = FragSourceTemplate.ExecuteTemplate(w, "fragSource", FragSourceData{
			SourceHTML: template.HTML(
				`<p class="b-state-frag-source-toolarge">File is too large to preview here. ` +
					`Open it in the source tab.</p>`,
			),
			PkgPath:     u.Path,
			File:        file,
			Line:        line,
			HeightParam: heightParam,
		})
		return http.StatusOK, nil
	}

	// Explicit span (StartLine..EndLine from the func/method node) wins;
	// otherwise center a ±fragSourceContextLines window on `line` per the
	// legacy fallback.
	var startLine int
	if endLine > 0 {
		startLine = line
	} else {
		startLine = line - fragSourceContextLines
		if startLine < 1 {
			startLine = 1
		}
		endLine = line + fragSourceContextLines
	}
	slice := sliceLines(content, startLine, endLine)

	var html template.HTML
	if h.deps.Highlighter != nil {
		rendered, err := h.deps.Highlighter.Render(file, slice)
		if err != nil {
			h.deps.Logger.Debug("frag=source highlight failed", "file", file, "err", err)
			html = htmlEscapePre(slice)
		} else {
			html = rendered
		}
	} else {
		html = htmlEscapePre(slice)
	}

	writeFragSuccessHeaders(w, height)
	w.WriteHeader(http.StatusOK)
	if err := FragSourceTemplate.ExecuteTemplate(w, "fragSource", FragSourceData{
		SourceHTML:  html,
		PkgPath:     u.Path,
		File:        file,
		Line:        line,
		HeightParam: heightParam,
	}); err != nil {
		h.deps.Logger.Error("frag=source template execute failed", "file", file, "err", err)
	}
	return http.StatusOK, nil
}

// writeFragError emits an HTTP-200 error fragment so htmx swaps a visible
// message instead of silently dropping a 4xx/5xx (ADR-004 §Decision §2).
// The Cache-Control: no-store header prevents nginx from caching transient
// failures.
func writeFragError(w http.ResponseWriter, message string, retryHints ...string) (int, *components.View) {
	var hint string
	if len(retryHints) > 0 {
		hint = retryHints[0]
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_ = FragErrorTemplate.ExecuteTemplate(w, "fragError", FragErrorData{
		Message:   message,
		RetryHint: hint,
	})
	return http.StatusOK, nil
}

// fragErrorFromClient maps a ClientAdapter error into the fragment-error
// pattern, hiding internal-error details from the client while logging
// the full error server-side. Always returns HTTP 200.
func (h *Handler) fragErrorFromClient(w http.ResponseWriter, err error, height int64, logKey, logVal string) (int, *components.View) {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		h.deps.Logger.Debug("fragment ctx ended", logKey, logVal, "err", err)
		return writeFragError(w, "Request timed out", "Please retry.")
	}
	status, msg := mapClientError(err, height)
	switch status {
	case http.StatusNotFound, http.StatusBadRequest, http.StatusRequestTimeout:
		return writeFragError(w, msg, "")
	default:
		h.deps.Logger.Error("fragment client error", logKey, logVal, "err", err, "status", status)
		return writeFragError(w, "Internal error", "Please retry.")
	}
}

// writeFragSuccessHeaders sets the canonical headers for a successful HTML
// fragment response: nosniff, noindex, and the parent page's cache strategy.
// Latest height → max-age=1 (collapses thundering herd at block tip); pinned
// → max-age=86400, immutable (hot audit path).
func writeFragSuccessHeaders(w http.ResponseWriter, height int64) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	if height > 0 {
		w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=1")
	}
}

// htmlEscapePre wraps the bytes in <pre>…</pre> with HTML escaping. Used
// as the fallback when Highlighter is nil or fails — keeps the fragment
// shape consistent for the controller-state hydration step.
func htmlEscapePre(b []byte) template.HTML {
	var buf bytes.Buffer
	buf.WriteString("<pre>")
	template.HTMLEscape(&buf, b)
	buf.WriteString("</pre>")
	return template.HTML(buf.String()) //nolint:gosec
}
