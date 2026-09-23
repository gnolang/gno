package omnisearch

import (
	"context"
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// chainSelectors are always registered: they cost the one query the package
// page already makes, and keep the omnibar useful with no indexer.
func chainSelectors() []*Selector {
	return []*Selector{
		renderSelector(),
		{
			Name:  "func",
			Hint:  "func:<name>",
			Label: "Functions",
			Scope: ScopePackage,
			resolve: func(ctx context.Context, h *Handler, q *Query, term string) ([]Result, error) {
				return h.resolveFuncs(ctx, q, term)
			},
		},
		{
			Name:  "type",
			Hint:  "type:<name>",
			Label: "Types",
			Scope: ScopePackage,
			resolve: func(ctx context.Context, h *Handler, q *Query, term string) ([]Result, error) {
				return h.resolveTypes(ctx, q, term)
			},
		},
		{
			Name:  "file",
			Hint:  "file:<name>",
			Label: "Files",
			Scope: ScopePackage,
			resolve: func(ctx context.Context, h *Handler, q *Query, term string) ([]Result, error) {
				return h.resolveFiles(ctx, q, term)
			},
		},
		{
			// Hands off to feature/state, which already owns `?state&search=`.
			Name:  "state",
			Hint:  "state:<name>",
			Label: "State",
			Scope: ScopePackage,
			resolve: func(ctx context.Context, h *Handler, q *Query, term string) ([]Result, error) {
				return []Result{{
					Title:  "Browse state matching " + term,
					Detail: "Opens the state explorer filtered to this name",
					Href:   stateSearchHref(q.PkgPath, term),
					Tags:   []string{"explorer"},
				}}, nil
			},
		},
		{
			Name:  "imports",
			Hint:  "imports",
			Label: "Imports",
			Scope: ScopePackage,
			Bare:  true,
			resolve: func(ctx context.Context, h *Handler, q *Query, term string) ([]Result, error) {
				return h.resolveImports(ctx, q, term)
			},
		},
	}
}

// webHref builds a `<pkgPath>$key&…` link through weburl's encoder. `?key=…`
// would land in u.Query, miss the WebQuery dispatch and route elsewhere.
func webHref(pkgPath string, wq url.Values) template.URL {
	if !isSafeRelPath(pkgPath) {
		return ""
	}
	u := weburl.GnoURL{Path: pkgPath, WebQuery: wq}
	return template.URL(u.EncodeWebURL()) //nolint:gosec // G203: path validated, webargs encoded by weburl.
}

// sourceHref deep-links a declaration site in the source view.
func sourceHref(pkgPath, file string, line int) template.URL {
	if file == "" || line <= 0 {
		return ""
	}
	href := fileHref(pkgPath, file)
	if href == "" {
		return ""
	}
	//nolint:gosec // G203: href is already validated and encoded; the line is an int.
	return href + template.URL("#L"+strconv.Itoa(line))
}

// stateSearchHref hands a name off to feature/state's own filtered view.
func stateSearchHref(pkgPath, term string) template.URL {
	return webHref(pkgPath, url.Values{"state": {""}, "search": {term}})
}

// actionHref links a callable realm function to its Action page.
func actionHref(pkgPath, fn string) template.URL {
	return webHref(pkgPath, url.Values{"help": {""}, "func": {fn}})
}

// fileHref links a source file in the source view.
func fileHref(pkgPath, file string) template.URL {
	return webHref(pkgPath, url.Values{"source": {""}, "file": {file}})
}

// synopsis is the first line of a doc comment, for the result's second row.
func synopsis(d string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(d), "\n")
	return line
}

// matches is a case-insensitive substring hit; an empty term matches all.
func matches(name, term string) bool {
	if term == "" {
		return true
	}
	return strings.Contains(strings.ToLower(name), strings.ToLower(term))
}

func (h *Handler) resolveFuncs(ctx context.Context, q *Query, term string) ([]Result, error) {
	jdoc, err := h.doc(ctx, q.PkgPath)
	if err != nil {
		return nil, fmt.Errorf("qdoc %s: %w", q.PkgPath, err)
	}

	out := make([]Result, 0, len(jdoc.Funcs))
	for _, fn := range jdoc.Funcs {
		if !matches(fn.Name, term) {
			continue
		}
		r := Result{
			Title:  fn.Signature,
			Detail: synopsis(fn.Doc),
			Href:   sourceHref(q.PkgPath, fn.File, fn.Line),
		}
		if fn.Crossing {
			r.Tags = append(r.Tags, "crossing")
		}
		if fn.Type != "" {
			r.Tags = append(r.Tags, "method on "+fn.Type)
		}
		// Only realms expose actions, and only on top-level funcs.
		if fn.Type == "" && fn.Name != "Render" && strings.HasPrefix(q.PkgPath, "/r/") {
			r.Tags = append(r.Tags, "action")
			if r.Href == "" {
				r.Href = actionHref(q.PkgPath, fn.Name)
			}
		}
		out = append(out, r)
	}
	return capResults(out), nil
}

func (h *Handler) resolveTypes(ctx context.Context, q *Query, term string) ([]Result, error) {
	jdoc, err := h.doc(ctx, q.PkgPath)
	if err != nil {
		return nil, fmt.Errorf("qdoc %s: %w", q.PkgPath, err)
	}

	out := make([]Result, 0, len(jdoc.Types))
	for _, t := range jdoc.Types {
		if !matches(t.Name, term) {
			continue
		}
		out = append(out, Result{
			Title:  "type " + t.Name,
			Detail: synopsis(t.Doc),
			Href:   sourceHref(q.PkgPath, t.File, t.Line),
			Tags:   typeTags(t),
		})
	}
	return capResults(out), nil
}

func typeTags(t *doc.JSONType) []string {
	tags := []string{}
	if t.Kind != "" {
		tags = append(tags, t.Kind)
	}
	if t.Alias {
		tags = append(tags, "alias")
	}
	return tags
}

func (h *Handler) resolveFiles(ctx context.Context, q *Query, term string) ([]Result, error) {
	files, err := h.deps.Client.ListFiles(ctx, q.PkgPath, 0)
	if err != nil {
		return nil, fmt.Errorf("list files %s: %w", q.PkgPath, err)
	}

	out := make([]Result, 0, len(files))
	for _, f := range files {
		if !matches(f, term) {
			continue
		}
		r := Result{
			Title: f,
			Href:  fileHref(q.PkgPath, f),
		}
		if strings.HasSuffix(f, "_test.gno") || strings.HasSuffix(f, "_filetest.gno") {
			r.Tags = append(r.Tags, "test")
		}
		out = append(out, r)
	}
	return capResults(out), nil
}

func (h *Handler) resolveImports(ctx context.Context, q *Query, term string) ([]Result, error) {
	jdoc, err := h.doc(ctx, q.PkgPath)
	if err != nil {
		return nil, fmt.Errorf("qdoc %s: %w", q.PkgPath, err)
	}

	out := make([]Result, 0, len(jdoc.Imports))
	for _, imp := range jdoc.Imports {
		if !matches(imp, term) {
			continue
		}
		kind, href := classifyImport(imp, h.deps.Domain)
		out = append(out, Result{
			Title: imp,
			Href:  href,
			Tags:  []string{kind},
		})
	}
	return capResults(out), nil
}

// classifyImport labels a dependency and links it when it is on this chain.
func classifyImport(p, domain string) (kind string, href template.URL) {
	switch {
	case strings.HasPrefix(p, domain+"/p/"):
		return "package", safePathHref(strings.TrimPrefix(p, domain))
	case strings.HasPrefix(p, domain+"/r/"):
		return "realm", safePathHref(strings.TrimPrefix(p, domain))
	case strings.Contains(p, "."):
		return "external", ""
	default:
		return "stdlib", ""
	}
}
