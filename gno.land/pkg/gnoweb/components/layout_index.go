package components

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"
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
	Title             string
	Description       string
	Canonical         string
	Image             string
	URL               string
	ChromaPath        string
	AssetsPath        string
	AnalyticsHostname string
	Remote            string
	ChainId           string
	BuildTime         string
}

// MaxBannerLength is the maximum character length for banner markdown source.
const MaxBannerLength = 400

// BannerVariant selects the banner's color scheme. Each value maps to an
// --s-color-bg-<variant>-default token that the design system already defines,
// so a banner never introduces a color of its own.
type BannerVariant string

const (
	BannerBrand   BannerVariant = "brand"
	BannerSuccess BannerVariant = "success"
	BannerInfo    BannerVariant = "info"
	BannerWarning BannerVariant = "warning"
	BannerCaution BannerVariant = "caution"
	BannerTip     BannerVariant = "tip"
	BannerNote    BannerVariant = "note"
)

// bannerVariantClass maps a variant to its CSS modifier class.
//
// The class names are spelled out in full on purpose. gnoweb's stylesheet is
// built with purgecss, which extracts whole tokens from *.go and *.html and
// drops every class it does not find. Assembling "b-banner--" + variant in the
// template would compile and pass tests, then ship an unstyled banner in
// production, because purgecss would never see the joined name. Keep the
// literals here, and keep this map as the only place they are written.
var bannerVariantClass = map[BannerVariant]string{
	BannerBrand:   "b-banner--brand",
	BannerSuccess: "b-banner--success",
	BannerInfo:    "b-banner--info",
	BannerWarning: "b-banner--warning",
	BannerCaution: "b-banner--caution",
	BannerTip:     "b-banner--tip",
	BannerNote:    "b-banner--note",
}

// ValidBannerVariant reports whether v names a known variant.
func ValidBannerVariant(v BannerVariant) bool {
	_, ok := bannerVariantClass[v]
	return ok
}

// BannerVariants returns every accepted variant name, sorted, for help text
// and error messages.
func BannerVariants() []string {
	out := make([]string, 0, len(bannerVariantClass))
	for v := range bannerVariantClass {
		out = append(out, string(v))
	}
	sort.Strings(out)
	return out
}

// bannerColorRe matches the colors a banner may carry: a CSS hex color, or a
// bare color keyword. Deliberately narrow, because the value is emitted into a
// style attribute as template.HTMLAttr: nothing that could close the attribute
// or open a second declaration is allowed through.
var bannerColorRe = regexp.MustCompile(`^(#([0-9a-fA-F]{3}|[0-9a-fA-F]{4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})|[a-zA-Z]{3,24})$`)

// ValidBannerColor reports whether c is an acceptable banner color.
func ValidBannerColor(c string) bool { return bannerColorRe.MatchString(c) }

// BannerOptions configures a banner's link target and appearance.
type BannerOptions struct {
	// URL, when it is an http or https URL, turns the whole banner into a
	// single clickable link and unwraps any inline markdown links to text.
	URL string
	// Variant selects a color scheme. Empty means BannerBrand.
	Variant BannerVariant
	// Color overrides the background with a literal CSS color, taking
	// precedence over Variant. Must satisfy ValidBannerColor. The text color
	// is not adjusted, so a light color makes the banner unreadable.
	Color string
}

// BannerData implements Component.
var _ Component = BannerData{}

// BannerData holds pre-rendered inline HTML from markdown.
type BannerData struct {
	content string
	url     string
	variant BannerVariant
	color   string // validated by bannerColorRe, empty when unset
}

func (b BannerData) Enabled() bool { return b.content != "" }
func (b BannerData) HasURL() bool  { return b.url != "" }
func (b BannerData) URL() string   { return b.url }

// VariantClass returns the CSS modifier class for this banner's variant.
func (b BannerData) VariantClass() string {
	if class, ok := bannerVariantClass[b.variant]; ok {
		return class
	}
	return bannerVariantClass[BannerBrand]
}

// Color returns the explicit background color set on this banner, or an empty
// string when the variant decides it.
//
// The template interpolates this inside a style attribute, so html/template
// escapes it in CSS context. bannerColorRe is defense in depth on top of that,
// and is what lets the CLI reject a bad value up front instead of quietly
// rendering a neutered one.
func (b BannerData) Color() string { return b.color }

func (b BannerData) Render(w io.Writer) (err error) {
	_, err = io.WriteString(w, b.content)
	return err
}

// NewBannerData parses inline markdown into a BannerData with pre-rendered HTML.
// Content after the first newline is discarded. Content is truncated to MaxBannerLength runes.
// If opts.URL is non-empty (http/https only), the banner acts as a single clickable link
// and any inline markdown links are unwrapped to plain text.
//
// An unknown opts.Variant or a malformed opts.Color is an error: a banner that
// silently ignores the color it was asked for is worse than one that refuses,
// because the operator only finds out by looking at the page.
func NewBannerData(markdown string, opts BannerOptions) (BannerData, error) {
	globalURL := opts.URL

	if opts.Variant != "" && !ValidBannerVariant(opts.Variant) {
		return BannerData{}, fmt.Errorf("unknown banner variant %q, want one of %s",
			opts.Variant, strings.Join(BannerVariants(), ", "))
	}
	if opts.Color != "" && !ValidBannerColor(opts.Color) {
		return BannerData{}, fmt.Errorf("invalid banner color %q, want a hex color or a CSS color keyword", opts.Color)
	}

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

	bd := BannerData{content: result, variant: opts.Variant, color: opts.Color}
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

	data.FooterData.Analytics.PageType = ClassifyPageType(data.Mode, data.BodyView.Type)
	data.FooterData.Analytics.Path = analyticsPath(data.HeaderData.RealmURL)
	data.FooterData.Analytics.ChainId = data.HeadData.ChainId
	data.FooterData.Analytics.AssetsPath = data.HeadData.AssetsPath
	data.FooterData.Analytics.BuildTime = data.HeadData.BuildTime
	data.FooterData.Analytics.Hostname = data.HeadData.AnalyticsHostname

	dataLayout := indexLayoutParams{
		IndexData: data,
		ViewType:  data.BodyView.String(),
		Theme:     data.Theme,
	}

	// Set dev mode based on view type and mode
	switch data.BodyView.Type {
	case HelpViewType, SourceViewType, DirectoryViewType, StatusViewType, StateViewType, OverviewViewType:
		dataLayout.IsDevmodView = true
	}

	return NewTemplateComponent("index", dataLayout)
}
