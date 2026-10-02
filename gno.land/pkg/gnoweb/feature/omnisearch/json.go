package omnisearch

import (
	"context"
	"encoding/json"
	"hash/fnv"
	"net/http"
	"strconv"
)

// jsonResult mirrors Result on the wire. Declared separately so the JSON
// contract is an explicit, reviewable surface rather than whatever the
// render struct happens to hold today.
type jsonResult struct {
	Title  string   `json:"title"`
	Detail string   `json:"detail,omitempty"`
	Href   string   `json:"href,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}

type jsonGroup struct {
	Label     string       `json:"label"`
	Source    Source       `json:"source"`
	Results   []jsonResult `json:"results"`
	Error     string       `json:"error,omitempty"`
	Truncated bool         `json:"truncated,omitempty"`
}

type jsonSelector struct {
	Name   string `json:"name"`
	Hint   string `json:"hint"`
	Label  string `json:"label"`
	Scope  Scope  `json:"scope"`
	Source Source `json:"source"`
	Bare   bool   `json:"bare,omitempty"`
}

type jsonIndexer struct {
	URL       string `json:"url"`
	LastBlock int    `json:"last_block,omitempty"`
	Error     string `json:"error,omitempty"`
}

type jsonResponse struct {
	Query string `json:"query"`
	// PkgPath echoes the scope so a client can tell a package-scoped answer
	// from a global one without re-parsing the URL it sent.
	PkgPath string `json:"pkg_path,omitempty"`
	// Selectors is returned on every response, not just on discovery: it is
	// how the omnibar learns which qualifiers this deployment can answer, and
	// a stale hint list is what produces a search box that offers something
	// the server will refuse.
	Selectors     []jsonSelector `json:"selectors"`
	Groups        []jsonGroup    `json:"groups,omitempty"`
	UnknownFilter string         `json:"unknown_filter,omitempty"`
	Indexer       *jsonIndexer   `json:"indexer,omitempty"`
}

// serveJSON answers the omnibar. The body is written here; the caller wraps
// the returned status with a nil view so no chrome is composed.
func (h *Handler) serveJSON(ctx context.Context, w http.ResponseWriter, r *http.Request, q *Query) int {
	data := h.build(ctx, q)

	resp := jsonResponse{
		Query:         data.Query,
		PkgPath:       data.PkgPath,
		Selectors:     make([]jsonSelector, 0, len(data.Selectors)),
		UnknownFilter: data.UnknownFilter,
	}
	for _, s := range data.Selectors {
		resp.Selectors = append(resp.Selectors, jsonSelector{
			Name: s.Name, Hint: s.Hint, Label: s.Label,
			Scope: s.Scope, Source: s.Source, Bare: s.Bare,
		})
	}
	for _, g := range data.Groups {
		jg := jsonGroup{
			Label: g.Label, Source: g.Source, Truncated: g.Truncated,
			Results: make([]jsonResult, 0, len(g.Results)),
		}
		if g.Err != nil {
			jg.Error = g.Err.Error()
		}
		for _, r := range g.Results {
			jg.Results = append(jg.Results, jsonResult{
				Title: r.Title, Detail: r.Detail, Href: string(r.Href), Tags: r.Tags,
			})
		}
		resp.Groups = append(resp.Groups, jg)
	}
	if data.Indexer != nil {
		ji := &jsonIndexer{URL: data.Indexer.URL, LastBlock: data.Indexer.LastBlock}
		if data.Indexer.Err != nil {
			ji.Error = data.Indexer.Err.Error()
		}
		resp.Indexer = ji
	}

	body, err := json.Marshal(resp)
	if err != nil {
		h.deps.Logger.Error("omnisearch: encode response", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return http.StatusInternalServerError
	}

	// The omnibar re-asks the same question constantly — retyping a
	// character, reopening the dropdown. An ETag turns a repeat into a 304
	// instead of a re-render, which max-age alone cannot do once it lapses.
	etag := `"` + strconv.FormatUint(bodyHash(body), 16) + `"`
	writeJSONHeaders(w)
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "max-age=1")
	if match := r.Header.Get("If-None-Match"); match == etag {
		w.WriteHeader(http.StatusNotModified)
		return http.StatusNotModified
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(body); err != nil {
		h.deps.Logger.Error("omnisearch: write response", "error", err)
	}
	return http.StatusOK
}

// bodyHash is FNV-1a: the ETag only has to change when the body does, so a
// non-cryptographic hash from the stdlib is the whole requirement.
func bodyHash(b []byte) uint64 {
	h := fnv.New64a()
	_, _ = h.Write(b)
	return h.Sum64()
}

// writeJSONError emits the same envelope shape as a successful response's
// failure fields, so a client has exactly one error format to handle.
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSONHeaders(w)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{Error: msg})
}

func writeJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
}
