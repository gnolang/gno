package ui

import "strings"

type DOM struct {
	// metadata
	Prefix       string
	Title        string
	WithComments bool

	// elements
	Header Element
	Body   Element
	Footer Element
}

func (dom DOM) String() string {
	output := ""

	if dom.Title != "" {
		output += H1(dom.Title).String(dom) + "\n"
	}

	if header := dom.Header.String(dom); header != "" {
		if dom.WithComments {
			output += "<!-- header -->"
		}
		output += header + "\n"
		if dom.WithComments {
			output += "<!-- /header -->"
		}
	}

	if body := dom.Body.String(dom); body != "" {
		if dom.WithComments {
			output += "<!-- body -->"
		}
		output += body + "\n"
		if dom.WithComments {
			output += "<!-- /body -->"
		}
	}

	if footer := dom.Footer.String(dom); footer != "" {
		if dom.WithComments {
			output += "<!-- footer -->"
		}
		output += footer + "\n"
		if dom.WithComments {
			output += "<!-- /footer -->"
		}
	}

	// TODO: cleanup double new-lines.

	return output
}

// XXX: rename Element to Div?
type Element []DomStringer

func (e *Element) Append(elems ...DomStringer) {
	*e = append(*e, elems...)
}

func (e *Element) String(dom DOM) string {
	output := ""
	for _, elem := range *e {
		output += elem.String(dom) + "\n"
	}
	return output
}

type Breadcrumb []DomStringer

func (b *Breadcrumb) Append(elems ...DomStringer) {
	*b = append(*b, elems...)
}

func (b Breadcrumb) String(dom DOM) string {
	output := ""
	for idx, entry := range b {
		if idx > 0 {
			output += " / "
		}
		output += entry.String(dom)
	}
	return output
}

type Link struct {
	Text string
	Path string
	URL  string
}

// TODO: image

// TODO: pager

func (l Link) String(dom DOM) string {
	url := ""
	switch {
	case l.Path != "" && l.URL != "":
		panic("a link should have a path or a URL, not both.")
	case l.Path != "":
		url = dom.Prefix + l.Path
	case l.URL != "":
		url = l.URL
	}

	return "[" + l.Text + "](" + url + ")"
}

type BulletList []DomStringer

func (bl BulletList) String(dom DOM) string {
	output := ""

	for _, entry := range bl {
		output += "- " + entry.String(dom) + "\n"
	}

	return output
}

func Text(s string) DomStringer {
	return Raw{Content: s}
}

type DomStringer interface {
	String(dom DOM) string
}

type Raw struct {
	Content string
}

func (r Raw) String(_ DOM) string {
	return r.Content
}

type (
	H1        string
	H2        string
	H3        string
	H4        string
	H5        string
	H6        string
	Bold      string
	Italic    string
	Code      string
	Paragraph string
	HR        struct{}
)

func (text H1) String(_ DOM) string        { return "# " + string(text) + "\n" }
func (text H2) String(_ DOM) string        { return "## " + string(text) + "\n" }
func (text H3) String(_ DOM) string        { return "### " + string(text) + "\n" }
func (text H4) String(_ DOM) string        { return "#### " + string(text) + "\n" }
func (text H5) String(_ DOM) string        { return "##### " + string(text) + "\n" }
func (text H6) String(_ DOM) string        { return "###### " + string(text) + "\n" }
func (text Bold) String(_ DOM) string      { return "**" + string(text) + "**" }
func (text Italic) String(_ DOM) string    { return "_" + string(text) + "_" }
func (text Paragraph) String(_ DOM) string { return "\n" + string(text) + "\n" }
func (_ HR) String(_ DOM) string           { return "\n---\n" }

func (text Code) String(_ DOM) string {
	// multiline
	if strings.Contains(string(text), "\n") {
		return "\n```\n" + string(text) + "\n```\n"
	}

	// single line
	return "`" + string(text) + "`"
}
