package gnoweb

import (
	"bytes"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/gnolang/gno/docs"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// DocsURLPrefix is the path under which the embedded repository docs are
// served. Sub-pages map to .md files inside the docs package, e.g.
// /docs/builders/getting-started -> builders/getting-started.md.
const DocsURLPrefix = "/docs"

// DocsGitHubBaseURL is the base URL used to resolve link targets that
// escape the docs/ tree (e.g. ../../examples/... or ../../gnovm/...). These
// are intentional cross-repo references in the source docs; we rewrite
// them to GitHub blob URLs so they remain clickable from gnoweb.
const DocsGitHubBaseURL = "https://github.com/gnolang/gno/blob/master/"

// docsLinkRE captures Markdown links and images: [label](target) and ![alt](src).
// Submatches: 1 = "!" or "", 2 = label/alt, 3 = target.
var docsLinkRE = regexp.MustCompile(`(!?)\[([^\]]*)\]\(([^)\s]+)\)`)

// DocsHandler serves the repository documentation embedded via the docs
// package. Markdown pages are rendered through the same goldmark pipeline
// as realms so the docs look consistent with the rest of the site and any
// renderer improvement applies to both. Images and other binary assets are
// served raw via http.FileServer.
//
// Minimal slice: no dedicated sidebar yet; README.md acts as the index and
// in-page TOC is provided by the realm view. Sidebar and admonition syntax
// extension are tracked as follow-ups.
type DocsHandler struct {
	Logger   *slog.Logger
	Static   StaticMetadata
	Renderer Renderer

	fsys      fs.FS
	assetsSrv http.Handler
}

// NewDocsHandler builds a DocsHandler backed by the embedded docs FS.
func NewDocsHandler(logger *slog.Logger, static StaticMetadata, renderer Renderer) *DocsHandler {
	fsys := docs.FS()
	return &DocsHandler{
		Logger:    logger,
		Static:    static,
		Renderer:  renderer,
		fsys:      fsys,
		assetsSrv: http.StripPrefix(DocsURLPrefix+"/", http.FileServer(http.FS(fsys))),
	}
}

func (h *DocsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rel := strings.TrimPrefix(r.URL.Path, DocsURLPrefix)
	rel = strings.TrimPrefix(rel, "/")

	// Pass through assets (images, embedded source samples).
	if strings.HasPrefix(rel, "images/") || strings.HasPrefix(rel, "_assets/") {
		h.assetsSrv.ServeHTTP(w, r)
		return
	}

	src, resolvedRel, ok := h.resolve(rel)
	if !ok {
		h.renderError(w, r, http.StatusNotFound, "page not found")
		return
	}

	// Transform Docusaurus-style :::kind admonitions into the GitHub
	// `> [!KIND]` blockquote syntax already handled by markdown/ext_alert.go,
	// then rewrite relative .md links to clean /docs/... URLs. Both passes
	// are pure text transforms and commute (line-prefix vs inline link).
	src = transformAdmonitions(src)
	src = rewriteDocsLinks(src, resolvedRel)

	// Theme cookie (mirrors HTTPHandler.Get to avoid FOUC).
	var theme string
	if c, err := r.Cookie("theme"); err == nil {
		if c.Value == "light" || c.Value == "dark" {
			theme = c.Value
		}
	}

	indexData := components.IndexData{
		HeadData: components.HeadData{
			AssetsPath: h.Static.AssetsPath,
			ChromaPath: h.Static.ChromaPath,
			ChainId:    h.Static.ChainId,
			Remote:     h.Static.RemoteHelp,
			BuildTime:  h.Static.BuildTime,
			Title:      h.Static.Domain + " - " + r.URL.Path,
		},
		FooterData: components.FooterData{
			Analytics:  h.Static.Analytics,
			AssetsPath: h.Static.AssetsPath,
			BuildTime:  h.Static.BuildTime,
		},
		Theme:  theme,
		Banner: h.Static.Banner,
		Mode:   components.ViewModeRealm,
	}

	// Synthetic GnoURL for breadcrumb / renderer context. The docs subtree
	// is not a realm path, so we hand-build the value rather than route it
	// through weburl.ParseFromURL (which validates against the realm path
	// regex).
	gnourl := &weburl.GnoURL{
		Path:   r.URL.Path,
		Domain: h.Static.Domain,
	}

	indexData.HeaderData = components.HeaderData{
		Breadcrumb: generateBreadcrumbPaths(gnourl),
		RealmURL:   *gnourl,
		ChainId:    h.Static.ChainId,
		Remote:     h.Static.RemoteHelp,
		Mode:       indexData.Mode,
		Static:     true,
	}

	var content bytes.Buffer
	if _, err := h.Renderer.RenderRealm(&content, gnourl, src, RealmRenderContext{
		ChainId: h.Static.ChainId,
		Remote:  h.Static.RemoteHelp,
		Domain:  h.Static.Domain,
	}); err != nil {
		h.Logger.Error("docs render failed", "path", r.URL.Path, "error", err)
		h.renderError(w, r, http.StatusInternalServerError, "render error")
		return
	}

	indexData.BodyView = components.DocsView(components.DocsData{
		ComponentContent: components.NewReaderComponent(&content),
		Sections:         buildSidebar(resolvedRel),
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := components.IndexLayout(indexData).Render(w); err != nil {
		h.Logger.Error("docs index layout render failed", "error", err)
	}
}

// resolve maps a relative URL path (without the /docs/ prefix) to embedded
// markdown bytes and the canonical relative file path used for link rewriting.
//
//	""                          -> README.md
//	"builders/getting-started"  -> builders/getting-started.md
//	                               or builders/getting-started/README.md
//	"foo.md"                    -> foo.md (direct)
func (h *DocsHandler) resolve(rel string) (src []byte, resolvedRel string, ok bool) {
	rel = path.Clean(rel)
	if rel == "." || rel == "" {
		b, ok := readFile(h.fsys, "README.md")
		return b, "README.md", ok
	}
	if strings.HasPrefix(rel, "..") {
		return nil, "", false
	}

	var candidates []string
	if strings.HasSuffix(rel, ".md") {
		candidates = []string{rel}
	} else {
		candidates = []string{rel + ".md", path.Join(rel, "README.md")}
	}
	for _, c := range candidates {
		if b, ok := readFile(h.fsys, c); ok {
			return b, c, true
		}
	}
	return nil, "", false
}

func (h *DocsHandler) renderError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	indexData := components.IndexData{
		HeadData: components.HeadData{
			AssetsPath: h.Static.AssetsPath,
			ChromaPath: h.Static.ChromaPath,
			ChainId:    h.Static.ChainId,
			Remote:     h.Static.RemoteHelp,
			BuildTime:  h.Static.BuildTime,
			Title:      h.Static.Domain + " - " + msg,
		},
		FooterData: components.FooterData{
			Analytics:  h.Static.Analytics,
			AssetsPath: h.Static.AssetsPath,
			BuildTime:  h.Static.BuildTime,
		},
		Banner:   h.Static.Banner,
		Mode:     components.ViewModeRealm,
		BodyView: components.StatusErrorComponent(msg),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := components.IndexLayout(indexData).Render(w); err != nil {
		h.Logger.Error("docs error layout render failed", "error", err)
	}
}

func readFile(fsys fs.FS, name string) ([]byte, bool) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, false
	}
	return b, true
}

// rewriteDocsLinks rewrites relative Markdown links so that they point at
// the gnoweb /docs/... URLs. Absolute URLs, anchors, and mailto: links are
// preserved. Images keep their extension; .md links get stripped to match
// the clean URL scheme served by DocsHandler.
//
// currentRel is the embed-relative file path of the document being rendered
// (e.g. "builders/getting-started.md"); link targets are resolved relative
// to its directory.
func rewriteDocsLinks(src []byte, currentRel string) []byte {
	base := path.Dir(currentRel)
	return docsLinkRE.ReplaceAllFunc(src, func(m []byte) []byte {
		sub := docsLinkRE.FindSubmatch(m)
		bang, label, target := sub[1], sub[2], string(sub[3])
		if shouldSkipLink(target) {
			return m
		}

		// Split off anchor / query.
		anchor := ""
		if i := strings.IndexAny(target, "#?"); i >= 0 {
			anchor = target[i:]
			target = target[:i]
		}
		if target == "" {
			return m
		}

		resolved := path.Clean(path.Join(base, target))

		var b bytes.Buffer
		b.Write(bang)
		b.WriteByte('[')
		b.Write(label)
		b.WriteString("](")
		if strings.HasPrefix(resolved, "..") {
			// Escapes the docs/ tree; resolve against the repo root on
			// GitHub so cross-repo references stay clickable.
			b.WriteString(DocsGitHubBaseURL)
			b.WriteString(strings.TrimPrefix(resolved, "../"))
		} else {
			// .md becomes clean URL; assets keep their extension.
			resolved = strings.TrimSuffix(resolved, ".md")
			b.WriteString(DocsURLPrefix)
			b.WriteByte('/')
			b.WriteString(resolved)
		}
		b.WriteString(anchor)
		b.WriteByte(')')
		return b.Bytes()
	})
}

func shouldSkipLink(target string) bool {
	switch {
	case target == "":
		return true
	case strings.HasPrefix(target, "http://"),
		strings.HasPrefix(target, "https://"),
		strings.HasPrefix(target, "//"),
		strings.HasPrefix(target, "/"),
		strings.HasPrefix(target, "#"),
		strings.HasPrefix(target, "mailto:"),
		strings.HasPrefix(target, "data:"):
		return true
	}
	return false
}

// buildSidebar adapts the docs package's parsed nav into the component's
// view model: internal item hrefs are turned into clean /docs/<path> URLs
// and the item matching the currently rendered page is flagged Active.
// currentRel is the embed-relative path of the current document, e.g.
// "builders/getting-started.md".
func buildSidebar(currentRel string) []components.DocsSidebarSection {
	parsed := docs.Sidebar()
	out := make([]components.DocsSidebarSection, 0, len(parsed))
	for _, sec := range parsed {
		viewSec := components.DocsSidebarSection{
			Title: sec.Title,
			Items: make([]components.DocsSidebarItem, 0, len(sec.Items)),
		}
		for _, it := range sec.Items {
			var href string
			active := false
			if it.External {
				href = it.Href
			} else {
				// it.Href is "builders/getting-started.md"; emit the
				// clean URL the handler serves.
				clean := strings.TrimSuffix(it.Href, ".md")
				href = DocsURLPrefix + "/" + clean
				active = it.Href == currentRel
			}
			viewSec.Items = append(viewSec.Items, components.DocsSidebarItem{
				Title:    it.Title,
				Href:     href,
				External: it.External,
				Active:   active,
			})
		}
		out = append(out, viewSec)
	}
	return out
}

// admonitionOpenRE matches a Docusaurus-style admonition opener:
//
//	:::info               -> kind=info,    title=""
//	:::tip Try this       -> kind=tip,     title="Try this"
//	:::warning Heads up!  -> kind=warning, title="Heads up!"
//
// Submatch 1 is the kind, submatch 2 is the (optional) inline title.
var admonitionOpenRE = regexp.MustCompile(`^:::(\w+)(?:\s+(.*))?$`)

// transformAdmonitions rewrites Docusaurus-style admonitions
//
//	:::kind [title]
//	body line 1
//	body line 2
//	:::
//
// into the GitHub-style alert blockquote already understood by
// markdown/ext_alert.go:
//
//	> [!KIND] title
//	> body line 1
//	> body line 2
//
// Lines inside fenced code blocks (``` or ~~~) are passed through
// unchanged, so ::: sequences in code samples are preserved verbatim.
// Unterminated admonitions are flushed at EOF.
func transformAdmonitions(src []byte) []byte {
	lines := strings.Split(string(src), "\n")
	var out strings.Builder
	out.Grow(len(src))

	var (
		inFence   bool
		fenceChar byte
		inAlert   bool
	)

	flushClose := func() {
		// nothing structurally to write; the blockquote ends on the first
		// non-`> ` line, which is what we emit after the closing ":::".
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track fenced code blocks so we don't mistake "::: " inside code
		// samples for admonitions.
		if !inAlert && (strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")) {
			ch := trimmed[0]
			if !inFence {
				inFence, fenceChar = true, ch
			} else if ch == fenceChar {
				inFence = false
			}
			out.WriteString(line)
			if i < len(lines)-1 {
				out.WriteByte('\n')
			}
			continue
		}
		if inFence {
			out.WriteString(line)
			if i < len(lines)-1 {
				out.WriteByte('\n')
			}
			continue
		}

		if !inAlert {
			if m := admonitionOpenRE.FindStringSubmatch(trimmed); m != nil {
				kind := strings.ToUpper(m[1])
				title := strings.TrimSpace(m[2])
				out.WriteString("> [!")
				out.WriteString(kind)
				out.WriteByte(']')
				if title != "" {
					out.WriteByte(' ')
					out.WriteString(title)
				}
				if i < len(lines)-1 {
					out.WriteByte('\n')
				}
				inAlert = true
				continue
			}
			out.WriteString(line)
			if i < len(lines)-1 {
				out.WriteByte('\n')
			}
			continue
		}

		// Inside an admonition.
		if trimmed == ":::" {
			flushClose()
			inAlert = false
			// Do not emit anything for the closer; the blockquote ends here.
			// Preserve the blank-line cadence so following content is parsed
			// as a separate block.
			if i < len(lines)-1 {
				out.WriteByte('\n')
			}
			continue
		}
		// Prefix with "> " (or just ">" for blank lines).
		if trimmed == "" {
			out.WriteByte('>')
		} else {
			out.WriteString("> ")
			out.WriteString(line)
		}
		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
	}

	return []byte(out.String())
}
