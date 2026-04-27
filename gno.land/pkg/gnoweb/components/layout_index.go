package components

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// ViewMode represents the current view mode of the application
// It affects the layout, navigation, and display of content
type ViewMode int

const (
	ViewModeExplorer ViewMode = iota // For exploring packages and paths
	ViewModeRealm                    // For realm content display
	ViewModePackage                  // For package content display
	ViewModeHome                     // For home page display
	ViewModeUser                     // For user page display
)

// View mode predicates
func (m ViewMode) IsExplorer() bool { return m == ViewModeExplorer }
func (m ViewMode) IsRealm() bool    { return m == ViewModeRealm }
func (m ViewMode) IsPackage() bool  { return m == ViewModePackage }
func (m ViewMode) IsUser() bool     { return m == ViewModeUser }
func (m ViewMode) IsHome() bool     { return m == ViewModeHome }

// ShouldShowDevTools returns whether dev tools should be shown for this mode
func (m ViewMode) ShouldShowDevTools() bool {
	return m != ViewModeHome
}

// ShouldShowGeneralLinks returns whether general navigation links should be shown
func (m ViewMode) ShouldShowGeneralLinks() bool {
	return m == ViewModeHome
}

type HeadData struct {
	Title       string
	Description string
	Canonical   string
	Image       string
	URL         string
	ChromaPath  string
	AssetsPath  string
	Analytics   bool
	Remote      string
	ChainId     string
	BuildTime   string
}

// MaxBannerLength is the maximum character length for banner markdown source.
const MaxBannerLength = 400

// BannerData implements Component.
var _ Component = BannerData{}

// BannerData holds pre-rendered inline HTML from markdown.
type BannerData struct {
	content string
	url     string
}

func (b BannerData) Enabled() bool { return b.content != "" }
func (b BannerData) HasURL() bool  { return b.url != "" }
func (b BannerData) URL() string   { return b.url }

func (b BannerData) Render(w io.Writer) (err error) {
	_, err = io.WriteString(w, b.content)
	return err
}

// NewBannerData parses inline markdown into a BannerData with pre-rendered HTML.
// Content after the first newline is discarded. Content is truncated to MaxBannerLength runes.
// If globalURL is non-empty (http/https only), the banner acts as a single clickable link
// and any inline markdown links are unwrapped to plain text.
func NewBannerData(markdown, globalURL string) (BannerData, error) {
	// Keep only the first line
	if i := strings.IndexAny(markdown, "\n\r"); i >= 0 {
		markdown = markdown[:i]
	}
	markdown = strings.TrimSpace(markdown)

	if markdown == "" {
		return BannerData{}, nil
	}

	// Truncate to max length (rune-safe)
	if runes := []rune(markdown); len(runes) > MaxBannerLength {
		markdown = string(runes[:MaxBannerLength])
	}

	// Validate global URL: only http/https allowed.
	globalURL = strings.TrimSpace(globalURL)
	hasGlobalURL := strings.HasPrefix(globalURL, "https://") || strings.HasPrefix(globalURL, "http://")

	md := goldmark.New(goldmark.WithExtensions(extension.Strikethrough))
	src := []byte(markdown)
	doc := md.Parser().Parse(text.NewReader(src))

	// Keep only Paragraph nodes (the inline-content wrapper). All other
	// block-level nodes (headings, code blocks, lists, HTML blocks, etc.)
	// are removed so the banner contains only inline markup.
	for c := doc.FirstChild(); c != nil; {
		next := c.NextSibling()
		if c.Kind() != ast.KindParagraph {
			doc.RemoveChild(doc, c)
		}
		c = next
	}

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Kind() != ast.KindLink {
			return ast.WalkContinue, nil
		}

		if hasGlobalURL {
			// Replace link node with its children (keep text, drop the <a>).
			parent := n.Parent()
			for c := n.FirstChild(); c != nil; {
				next := c.NextSibling()
				parent.InsertBefore(parent, n, c)
				c = next
			}
			parent.RemoveChild(parent, n)
			return ast.WalkSkipChildren, nil
		}

		n.SetAttributeString("target", "_blank")
		n.SetAttributeString("rel", "noopener noreferrer")
		return ast.WalkContinue, nil
	})

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, src, doc); err != nil {
		return BannerData{}, fmt.Errorf("banner markdown rendering: %w", err)
	}

	// Strip the <p></p> wrapper that goldmark adds for single-paragraph content.
	result := strings.TrimSpace(buf.String())
	if after, ok := strings.CutPrefix(result, "<p>"); ok {
		if inner, ok := strings.CutSuffix(after, "</p>"); ok {
			result = inner
		}
	}

	bd := BannerData{content: result}
	if hasGlobalURL {
		bd.url = globalURL
	}
	return bd, nil
}

type IndexData struct {
	HeadData
	HeaderData
	FooterData
	BodyView *View
	Mode     ViewMode
	Theme    string
	Banner   BannerData
}

type indexLayoutParams struct {
	IndexData

	// Additional data
	IsDevmodView bool
	ViewType     string
	JSController string
	Theme        string
}

func IndexLayout(data IndexData) Component {
	data.FooterData = EnrichFooterData(data.FooterData)
	data.HeaderData = EnrichHeaderData(data.HeaderData, data.Mode)

	dataLayout := indexLayoutParams{
		IndexData: data,
		ViewType:  data.BodyView.String(),
		Theme:     data.Theme,
	}

	// Set dev mode based on view type and mode
	switch data.BodyView.Type {
	case HelpViewType, SourceViewType, DirectoryViewType, StatusViewType:
		dataLayout.IsDevmodView = true
	}

	return NewTemplateComponent("index", dataLayout)
}
