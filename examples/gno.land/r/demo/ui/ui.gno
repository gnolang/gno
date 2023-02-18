package ui

import "gno.land/p/demo/ui"

func Render(path string) string {
	// TODO: pagination
	// TODO: non-standard markdown
	// TODO: error, warn
	// TODO: header
	// TODO: HTML
	// TODO: toc

	var dom ui.DOM

	dom.Title = "UI Demo"

	dom.Header.Append(ui.Breadcrumb(
		ui.Link{Text: "foo", Path: "foo"},
		ui.Link{Text: "bar", Path: "foo/bar"},
	))

	dom.Body.Append(
		ui.Paragraph("Simple UI demonstration."),
		ui.BulletList(
			ui.Text("a text"),
			ui.Link{Text: "a relative link", Path: "foobar"},
			ui.Text("another text"),
			// ui.H1("a H1 text"),
			ui.Bold("a bold text"),
			ui.Italic("italic text"),
			ui.Text("raw markdown with **bold** text in the middle."),
			ui.Code("some inline code"),
			ui.Link{Text: "a remote link", URL: "https://gno.land"},
		),
	)

	dom.Footer.Append(ui.Text("I'm the footer."))
	dom.Body.Append(ui.Text("another string."))
	dom.Body.Append(ui.Paragraph("a paragraph."), ui.HR())

	return dom.String()
}
