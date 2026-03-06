package markdown

import (
	"errors"
	"net/url"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Error messages for invalid link formats
var ErrLinkInvalidURL = errors.New("invalid URL format")

const (
	// Tooltips info for link types
	tooltipExternalLink = "External link"
	tooltipInternalLink = "Cross package link"
	tooltipTxLink       = "Transaction link"
	tooltipUserLink     = "User profile"

	// SVG icon ids for link types
	iconExternalLink = "ico-external-link"
	iconInternalLink = "ico-internal-link"
	iconTxLink       = "ico-tx-link"
	iconUserLink     = "ico-user-link"

	// CSS classes for link types
	classLinkExternal = "link-external"
	classLinkInternal = "link-internal"
	classLinkTx       = "link-tx"
	classLinkUser     = "link-user"
)

// GnoLinkType represents the type of a link
type GnoLinkType int

const (
	GnoLinkTypeInvalid GnoLinkType = iota
	GnoLinkTypeExternal
	GnoLinkTypePackage
	GnoLinkTypeInternal
	GnoLinkTypeUser
)

func (t GnoLinkType) String() string {
	switch t {
	case GnoLinkTypeExternal:
		return "external"
	case GnoLinkTypePackage:
		return "package"
	case GnoLinkTypeInternal:
		return "internal"
	case GnoLinkTypeUser:
		return "user"
	}
	return "unknown"
}

var KindGnoLink = ast.NewNodeKind("GnoLink")

// GnoLink represents a link with Gno-specific metadata
type GnoLink struct {
	*ast.Link
	LinkType GnoLinkType
	GnoURL   *weburl.GnoURL
}

func (n *GnoLink) Dump(source []byte, level int) {
	m := map[string]string{}
	m["Destination"] = string(n.Destination)
	m["Title"] = string(n.Title)
	m["LinkType"] = n.LinkType.String()
	ast.DumpHelper(n, source, level, m, nil)
}

// Kind implements Node.Kind.
func (*GnoLink) Kind() ast.NodeKind {
	return KindGnoLink
}

// linkTransformer implements ASTTransformer
type linkTransformer struct{}

// Transform replaces ast.Link nodes with GnoLink nodes in two passes.
func (t *linkTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	orig, ok := getUrlFromContext(pc)
	if !ok {
		return
	}

	// Traverse through the document and transform link nodes to GnoLink nodes.
	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		link, ok := node.(*ast.Link)
		if !ok {
			return ast.WalkContinue, nil
		}

		// Create a new GnoLink node wrapping the original link.
		gnoLink := &GnoLink{Link: link}

		// Replace the original link with the GnoLink wrapper.
		parent, next := node.Parent(), node.NextSibling()
		parent.RemoveChild(parent, node)
		parent.InsertBefore(parent, next, gnoLink)

		// Parse destination URL and check for validity.
		dest, err := url.Parse(string(link.Destination))
		if err != nil {
			gnoLink.LinkType = GnoLinkTypeInvalid
			return ast.WalkContinue, nil
		}

		// Detect and set the GnoLink type.
		gnoLink.GnoURL, gnoLink.LinkType = detectLinkType(dest, orig)

		return ast.WalkContinue, nil
	})
}

// detectLinkType detects the type of link based on the destination
func detectLinkType(dest *url.URL, orig *weburl.GnoURL) (*weburl.GnoURL, GnoLinkType) {
	// Attempt to parse the destination as a GnoURL.
	target, err := weburl.ParseFromURL(dest)
	if err != nil {
		if dest.Scheme == "" {
			// If there's no scheme, consider it as a relative path.
			return nil, GnoLinkTypePackage
		}

		// Otherwise, treat it as an external URL.
		return nil, GnoLinkTypeExternal
	}

	// Check if it's a user link first
	if target.IsUser() {
		return target, GnoLinkTypeUser
	}

	// Extract domain and namespace from the target.
	targetDomain := target.Domain
	targetName := target.Namespace()

	switch {
	case targetDomain != "" && targetDomain != orig.Domain:
		// External: the domain does not match the origin's domain.
		return target, GnoLinkTypeExternal
	case targetName != "" && targetName == orig.Namespace():
		// Package: the namespace matches the origin's namespace.
		return target, GnoLinkTypePackage
	default:
		// Internal: it's neither external nor a package link.
		return target, GnoLinkTypeInternal
	}
}

// linkRenderer implements NodeRenderer
type linkRenderer struct{}

// RegisterFuncs registers the renderer functions
func (r *linkRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindGnoLink, r.renderGnoLink)
}

// attr represents an HTML attribute
type attr struct {
	name  string
	value string
}

// renderStringAttributes writes an HTML attribute.
// variant of html.RenderAttributes with custom attributes
// XXX: We probably want this as a general helper for futur extension.
func renderStringAttributes(w util.BufWriter, attrs []attr) {
	for _, attr := range attrs {
		w.WriteByte(' ')
		w.WriteString(attr.name)
		w.WriteString(`="`)
		w.Write(util.EscapeHTML(util.StringToReadOnlyBytes(attr.value)))
		w.WriteByte('"')
	}
}

// linkTypeInfo contains information about a link type.
type linkTypeInfo struct {
	tooltip string
	iconID  string
	class   string
}

// getLinkIcons returns all icons that should be displayed for a given link
func getLinkIcons(n *GnoLink) []linkTypeInfo {
	var icons []linkTypeInfo

	// Add type-specific icon (external/internal)
	if n.LinkType != GnoLinkTypePackage {
		switch n.LinkType {
		case GnoLinkTypeExternal:
			icons = append(icons, linkTypeInfo{tooltipExternalLink, iconExternalLink, classLinkExternal})
		case GnoLinkTypeInternal:
			icons = append(icons, linkTypeInfo{tooltipInternalLink, iconInternalLink, classLinkInternal})
		case GnoLinkTypeUser:
			icons = append(icons, linkTypeInfo{tooltipUserLink, iconUserLink, classLinkUser})
		}
	}

	// Add Tx icon for non-external links with help webquery
	if n.LinkType != GnoLinkTypeExternal && n.GnoURL != nil && n.GnoURL.WebQuery.Has("help") {
		icons = append(icons, linkTypeInfo{tooltipTxLink, iconTxLink, classLinkTx})
	}

	return icons
}

// renderGnoLink renders a link node.
func (r *linkRenderer) renderGnoLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n, ok := node.(*GnoLink)
	if !ok {
		return ast.WalkContinue, nil
	}

	if n.LinkType == GnoLinkTypeInvalid {
		if entering {
			w.WriteString("<!-- invalid link -->")
		}
		return ast.WalkSkipChildren, nil
	}

	if entering {
		w.WriteString(`<a href="`)
		if !html.IsDangerousURL(n.Destination) {
			w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		}
		w.WriteByte('"')

		// Prepare additional link attributes.
		attrs := []attr{}
		if n.LinkType == GnoLinkTypeExternal {
			attrs = append(attrs, attr{"rel", "noopener nofollow ugc"})
		}
		if n.Title != nil {
			attrs = append(attrs, attr{"title", string(n.Title)})
		}

		// Render additional attributes
		renderStringAttributes(w, attrs)

		// Close tag and continue
		w.WriteByte('>')
		return ast.WalkContinue, nil
	}

	// Render all icons dynamically
	for _, icon := range getLinkIcons(n) {
		w.WriteString("<span")
		renderStringAttributes(w, []attr{
			{"class", icon.class + " tooltip"},
			{"data-tooltip-target", "info"},
			{"data-tooltip", icon.tooltip},
			{"title", icon.tooltip},
		})
		w.WriteByte('>')
		w.WriteString(`<svg class="c-icon"><use href="#` + icon.iconID + `"></use></svg>`)
		w.WriteString("</span>")
	}

	// Write closing tag <a>.
	w.WriteString("</a>")

	return ast.WalkContinue, nil
}

// linkExtension is a Goldmark extension that handles link rendering with special attributes
// for external, internal, and same-package links.
type linkExtension struct{}

// ExtLinks instance for extending markdown with link functionality
var ExtLinks = &linkExtension{}

// Extend adds the LinkExtension to the provided Goldmark markdown processor
func (l *linkExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&linkTransformer{}, 500),
	))

	// Register our renderer with a higher priority than the default renderer
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&linkRenderer{}, 500),
	))
}
