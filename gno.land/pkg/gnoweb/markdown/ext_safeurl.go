package markdown

import (
	"context"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/safeurl"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// SafetyStatus re-exports safeurl.SafetyStatus for use in markdown package.
type SafetyStatus = safeurl.SafetyStatus

// Safety status constants.
const (
	StatusUnknown     = safeurl.StatusUnknown
	StatusSafe        = safeurl.StatusSafe
	StatusUnsafe      = safeurl.StatusUnsafe
	StatusUnavailable = safeurl.StatusUnavailable
)

// Context key for storing safety results.
var safeURLResultsKey = parser.NewContextKey()

// getSafetyResultsFromContext retrieves safety scan results from the parser context.
func getSafetyResultsFromContext(ctx parser.Context) (map[string]safeurl.ScanResult, bool) {
	results, ok := ctx.Get(safeURLResultsKey).(map[string]safeurl.ScanResult)
	return results, ok
}

// KindSafeImage is the node kind for images with safety metadata.
var KindSafeImage = ast.NewNodeKind("SafeImage")

// SafeImage wraps ast.Image with safety status.
type SafeImage struct {
	*ast.Image
	SafetyStatus SafetyStatus
	Verdict      string
}

// Kind implements ast.Node.
func (*SafeImage) Kind() ast.NodeKind {
	return KindSafeImage
}

// Dump implements ast.Node.
func (n *SafeImage) Dump(source []byte, level int) {
	m := map[string]string{
		"Destination":  string(n.Destination),
		"Title":        string(n.Title),
		"SafetyStatus": n.SafetyStatus.String(),
	}
	ast.DumpHelper(n, source, level, m, nil)
}

// safeURLTransformer implements ASTTransformer.
// It runs at priority 400 (before linkTransformer at 500) to collect and scan
// all external URLs, then annotates the AST nodes with safety status.
type safeURLTransformer struct {
	validator *safeurl.Validator
}

// Transform collects all external URLs, scans them, and updates AST nodes.
func (t *safeURLTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	if t.validator == nil || !t.validator.IsEnabled() {
		return
	}

	// Phase 1: Collect all external URLs from links and images
	var urls []string
	urlSet := make(map[string]bool)

	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		var dest string
		switch n := node.(type) {
		case *ast.Link:
			dest = string(n.Destination)
		case *ast.Image:
			dest = string(n.Destination)
		default:
			return ast.WalkContinue, nil
		}

		if dest != "" && isExternalURL(dest) && !urlSet[dest] {
			urls = append(urls, dest)
			urlSet[dest] = true
		}

		return ast.WalkContinue, nil
	})

	if len(urls) == 0 {
		return
	}

	// Phase 2: Validate all URLs in batch
	ctx := context.Background()
	results := t.validator.ValidateURLs(ctx, urls)

	// Store results in parser context for link transformer to use
	pc.Set(safeURLResultsKey, results)

	// Phase 3: Wrap Image nodes in SafeImage with safety status
	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		img, ok := node.(*ast.Image)
		if !ok {
			return ast.WalkContinue, nil
		}

		dest := string(img.Destination)
		if !isExternalURL(dest) {
			return ast.WalkContinue, nil
		}

		result, ok := results[dest]
		if !ok {
			return ast.WalkContinue, nil
		}

		// Create SafeImage wrapper
		safeImg := &SafeImage{
			Image:        img,
			SafetyStatus: result.Status,
			Verdict:      result.Verdict,
		}

		// Replace Image with SafeImage in the tree
		parent := img.Parent()
		if parent != nil {
			parent.ReplaceChild(parent, img, safeImg)
		}

		return ast.WalkContinue, nil
	})
}

// isExternalURL checks if a URL is external (requires safety validation).
func isExternalURL(url string) bool {
	// Empty or anchor-only URLs are internal
	if url == "" || strings.HasPrefix(url, "#") {
		return false
	}

	// Relative URLs are internal
	if strings.HasPrefix(url, "/") && !strings.HasPrefix(url, "//") {
		return false
	}

	// Data URIs don't need external validation
	if strings.HasPrefix(url, "data:") {
		return false
	}

	// Check for scheme
	if strings.Contains(url, "://") {
		// gno.land URLs are internal
		lowerURL := strings.ToLower(url)
		if strings.Contains(lowerURL, "gno.land") {
			return false
		}
		return true
	}

	// Protocol-relative URLs (//example.com) are external
	if strings.HasPrefix(url, "//") {
		return true
	}

	// No scheme - could be relative
	return false
}

// safeImageRenderer renders SafeImage nodes with safety-aware HTML.
type safeImageRenderer struct {
	html.Config
}

// RegisterFuncs registers the SafeImage renderer.
func (r *safeImageRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindSafeImage, r.renderSafeImage)
}

// renderSafeImage renders a SafeImage node based on its safety status.
func (r *safeImageRenderer) renderSafeImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*SafeImage)

	switch n.SafetyStatus {
	case StatusUnavailable:
		// Render as plain text URL (not clickable)
		w.WriteString(`<span class="img-unavailable" title="Unable to verify image safety">[Image: `)
		w.Write(util.EscapeHTML(n.Destination))
		w.WriteString(`]</span>`)
		return ast.WalkSkipChildren, nil

	case StatusUnsafe:
		// Render with warning overlay
		w.WriteString(`<span class="img-unsafe">`)
		w.WriteString(`<img src="`)
		if !html.IsDangerousURL(n.Destination) {
			w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		}
		w.WriteString(`"`)

		// Alt text
		w.WriteString(` alt="`)
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if t, ok := c.(*ast.Text); ok {
				w.Write(util.EscapeHTML(t.Segment.Value(source)))
			}
		}
		w.WriteString(`"`)

		// Title if present
		if n.Title != nil {
			w.WriteString(` title="`)
			w.Write(util.EscapeHTML(n.Title))
			w.WriteString(`"`)
		}

		w.WriteString(` />`)
		w.WriteString(`<span class="img-warning tooltip" data-tooltip="This image may be unsafe" title="This image may be unsafe">`)
		w.WriteString(`<svg class="c-icon"><use href="#ico-warning"></use></svg>`)
		w.WriteString(`</span>`)
		w.WriteString(`</span>`)
		return ast.WalkSkipChildren, nil

	default:
		// Safe or unknown - render with safety info
		w.WriteString(`<img src="`)
		if !html.IsDangerousURL(n.Destination) {
			w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		}
		w.WriteString(`"`)

		// Alt text
		w.WriteString(` alt="`)
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if t, ok := c.(*ast.Text); ok {
				w.Write(util.EscapeHTML(t.Segment.Value(source)))
			}
		}
		w.WriteString(`"`)

		// Title with safety info
		title := ""
		if n.Title != nil {
			title = string(n.Title)
		}
		if n.SafetyStatus == StatusSafe && n.Verdict != "" {
			if title != "" {
				title += " | "
			}
			title += "SafeURL: " + n.Verdict
		}
		if title != "" {
			w.WriteString(` title="`)
			w.Write(util.EscapeHTML([]byte(title)))
			w.WriteString(`"`)
		}

		// Add data attribute for safety status
		if n.SafetyStatus == StatusSafe {
			w.WriteString(` data-safeurl-status="safe"`)
			if n.Verdict != "" {
				w.WriteString(` data-safeurl-verdict="`)
				w.Write(util.EscapeHTML([]byte(n.Verdict)))
				w.WriteString(`"`)
			}
		}

		w.WriteString(` />`)
		return ast.WalkSkipChildren, nil
	}
}

// safeURLExtension is a Goldmark extension that adds URL safety validation.
type safeURLExtension struct{}

// ExtSafeURL is the SafeURL extension instance.
var ExtSafeURL = &safeURLExtension{}

// Extend adds the SafeURL extension to the provided Goldmark markdown processor.
func (e *safeURLExtension) Extend(m goldmark.Markdown, validator *safeurl.Validator) {
	// Add transformer at priority 400 (before link transformer at 500)
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&safeURLTransformer{validator: validator}, 400),
	))

	// Add SafeImage renderer at priority 400
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&safeImageRenderer{}, 400),
	))
}
