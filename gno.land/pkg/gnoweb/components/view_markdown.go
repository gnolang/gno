package components

import "bytes"

// MarkdownViewType marks a View whose content is raw markdown, to be served
// verbatim as text/markdown without the HTML page layout.
const MarkdownViewType ViewType = "markdown-view"

// MarkdownView returns a View that renders the given content as-is. The handler
// layer recognizes this view type and serves it with a text/markdown
// Content-Type, bypassing IndexLayout. The view reads content without copying
// it, so the caller must not mutate the slice until the view is rendered.
func MarkdownView(content []byte) *View {
	return &View{
		Type:      MarkdownViewType,
		Component: NewReaderComponent(bytes.NewReader(content)),
	}
}
