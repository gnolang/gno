package ui

type DOM struct {
	// metadata
	Prefix string
	Title  string

	// elements
	Header Element
	Body   Element
	Footer Element
}

func (dom DOM) String() string {
	output := ""

	if dom.Title != "" {
		output += H1(dom.Title).String() + "\n"
	}

	if !dom.Header.Empty() {
		output += dom.Header.Content + "\n"
	}

	if !dom.Body.Empty() {
		output += dom.Body.Content + "\n"
	}

	if !dom.Footer.Empty() {
		output += dom.Footer.Content + "\n"
	}

	// TODO: cleanup double new-lines.

	return output
}

type Element struct {
	Content string
}

func (e Element) Empty() bool {
	return e.Content == ""
}

func (e *Element) Append(elems ...stringer) {
	for _, elem := range elems {
		e.Content += elem.String() + "\n"
	}
}

func Breadcrumb(entries ...stringer) stringer {
	output := ""
	for idx, entry := range entries {
		if idx > 0 {
			output += " / "
		}
		output += entry.String()
	}
	return Raw{Content: output}
}

type Link struct {
	Text string
	Path string
	URL  string
}

// TODO: image

// TODO: pager

func (l Link) String() string {
	// TODO: additional arguments.
	if l.Path != "" && l.URL != "" {
		panic("a link should have a path OR a URL.")
	}

	if l.Path != "" {
		// TODO: prefix
		l.URL = "prefix" + l.Path
	}
	return "[" + l.Text + "](" + l.URL + ")"
}

func Paragraph(s string) stringer {
	return Raw{Content: "\n" + s + "\n"}
}

func BulletList(entries ...stringer) stringer {
	output := ""

	for _, entry := range entries {
		output += "- " + entry.String() + "\n"
	}

	return Raw{Content: output}
}

func Text(s string) stringer {
	return Raw{Content: s}
}

type stringer interface {
	String() string
}

type Raw struct {
	Content string
}

func (r Raw) String() string {
	return r.Content
}

func H1(text string) stringer     { return Raw{Content: "# " + text + "\n"} }
func H2(text string) stringer     { return Raw{Content: "## " + text + "\n"} }
func H3(text string) stringer     { return Raw{Content: "### " + text + "\n"} }
func H4(text string) stringer     { return Raw{Content: "#### " + text + "\n"} }
func H5(text string) stringer     { return Raw{Content: "##### " + text + "\n"} }
func H6(text string) stringer     { return Raw{Content: "###### " + text + "\n"} }
func Bold(text string) stringer   { return Raw{Content: "**" + text + "**"} }
func Italic(text string) stringer { return Raw{Content: "_" + text + "_"} }
func Code(text string) stringer   { return Raw{Content: "`" + text + "`"} }
func HR() stringer                { return Raw{Content: "\n---\n"} }
