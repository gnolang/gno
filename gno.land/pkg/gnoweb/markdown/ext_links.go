package markdown

import (
	"errors"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

const (
	// Tooltips info for link types
	tooltipExternalLink = "External link"
	tooltipInternalLink = "Cross package link"
	tooltipTxLink       = "Transaction link"

	// Icons for link types
	iconExternalLink = "↗"
	iconInternalLink = "↔"
	iconTxLink       = "⚡︎"

	// CSS classes for link types
	classLinkExternal = "link-external"
	classLinkInternal = "link-internal"
	classLinkTx       = "link-tx"
)

// LinkExtension is a Goldmark extension that handles link rendering with special attributes
// for external, internal, and same-package links.
type LinkExtension struct{}

// linkRenderer implements NodeRenderer
type linkRenderer struct {
	domain string
	path   string
}

// linkTypeInfo contains information about a link type
type linkTypeInfo struct {
	tooltip string
	icon    string
	class   string
}

// attr represents an HTML attribute
type attr struct {
	name  string
	value string
}

// RegisterFuncs registers the renderer functions
func (r *linkRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindLink, r.renderLink)
}

// isSamePackage checks if the link points to the same package
// - Link is an anchor (starts with #)
// - Link is relative (no leading / and no protocol -> root link)
// - Link points to the same package in /r/ path
// - Link points to the same package in /p/ path
// - Link points to the same package in /u/ path
func isSamePackage(dest, pathWithoutR string) bool {
	return strings.HasPrefix(dest, "#") ||
		(!strings.HasPrefix(dest, "/") && !strings.Contains(dest, "://")) ||
		strings.Contains(dest, "/r/"+pathWithoutR) ||
		strings.Contains(dest, "/p/"+pathWithoutR) ||
		strings.Contains(dest, "/u/"+pathWithoutR)
}

// Error messages for invalid link formats
var (
	ErrLinkInvalidURL = errors.New("invalid URL format")
)

// writeHTMLAttr writes an HTML attribute
func writeHTMLAttr(w util.BufWriter, name, value string) {
	_, _ = w.WriteString(" ")
	_, _ = w.WriteString(name)
	_, _ = w.WriteString(`="`)
	_, _ = w.WriteString(value)
	_, _ = w.WriteString(`"`)
}

// writeHTMLTag writes an HTML tag with its attributes
func writeHTMLTag(w util.BufWriter, tag string, attrs []attr) {
	_, _ = w.WriteString("<" + tag)
	for _, a := range attrs {
		writeHTMLAttr(w, a.name, a.value)
	}
	_, _ = w.WriteString(">")
}

// detectLinkType detects the type of link based on the destination
func detectLinkType(dest, domain, path string) (string, bool, error) {
	// Extract the package name from the path (e.g., "r/test/foo" -> "test")
	pathWithoutR := strings.TrimPrefix(path, "r/")
	if idx := strings.Index(pathWithoutR, "/"); idx != -1 {
		pathWithoutR = pathWithoutR[:idx]
	}

	// Check if the link is external:
	// - Contains a protocol (e.g., http://, https://)
	// - Contains a domain different from the current one
	if strings.Contains(dest, "://") && !strings.Contains(dest, "://"+domain) {
		return "external", false, nil
	}

	// Check if the link is a package link
	if isSamePackage(dest, pathWithoutR) {
		return "package", strings.Contains(dest, "$help"), nil
	}

	// All other links are internal
	return "internal", strings.Contains(dest, "$help"), nil
}

var linkTypes = map[string]linkTypeInfo{
	"external": {tooltipExternalLink, iconExternalLink, classLinkExternal},
	"internal": {tooltipInternalLink, iconInternalLink, classLinkInternal},
	"tx":       {tooltipTxLink, iconTxLink, classLinkTx},
}

// renderLink renders a link node
func (r *linkRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	dest := string(n.Destination)
	linkType, hasHelp, err := detectLinkType(dest, r.domain, r.path)
	if err != nil {
		fmt.Fprintf(w, "<!-- link error: %s -->\n", err.Error())
		return ast.WalkContinue, nil
	}

	if !entering {
		// Add the Tx icon span if needed
		if hasHelp {
			txAttrs := []attr{
				{"class", classLinkTx + " tooltip"},
				{"data-tooltip", tooltipTxLink},
			}
			writeHTMLTag(w, "span", txAttrs)
			_, _ = w.WriteString(iconTxLink)
			_, _ = w.WriteString("</span>")
		}

		// Add external/internal icon span if needed
		if linkType != "package" {
			if info, ok := linkTypes[linkType]; ok {
				attrs := []attr{
					{"class", info.class + " tooltip"},
					{"data-tooltip", info.tooltip},
				}
				writeHTMLTag(w, "span", attrs)
				_, _ = w.WriteString(info.icon)
				_, _ = w.WriteString("</span>")
			}
		}

		_, _ = w.WriteString("</a>")
		return ast.WalkContinue, nil
	}

	// Prepare link attributes with href first
	attrs := []attr{{"href", string(n.Destination)}}
	if linkType == "external" {
		attrs = append(attrs, attr{"rel", "noopener nofollow ugc"})
	}
	if n.Title != nil {
		attrs = append(attrs, attr{"title", string(n.Title)})
	}

	// Write opening <a> tag
	writeHTMLTag(w, "a", attrs)

	return ast.WalkContinue, nil
}

// Extend adds the LinkExtension to the provided Goldmark markdown processor
func (l *LinkExtension) Extend(m goldmark.Markdown, domain, path string) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&linkRenderer{
			domain: domain,
			path:   path,
		}, 500),
	))
}

// Links instance for extending markdown with link functionality
var Links = &LinkExtension{}
