package home

import (
	"strings"

	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/sunspirit/md/v0"
)

func Render(path string) string {
	var sb strings.Builder

	sb.WriteString(md.H1("Sunspirit's Home") + md.LineBreak(1))

	sb.WriteString(md.Paragraph(ufmt.Sprintf(
		"Welcome to Sunspirit’s home! This is where I’ll bring %s to Gno.land, crafted with my experience and creativity.",
		md.Italic(md.Bold("simple, useful dapps")),
	)) + md.LineBreak(1))

	sb.WriteString(md.Paragraph(ufmt.Sprintf(
		"📚 I’ve created a Markdown rendering library at %s. Feel free to use it for your own projects!",
		md.Link("gno.land/p/sunspirit/md/v0", "/p/sunspirit/md/v0"),
	)) + md.LineBreak(1))

	sb.WriteString(md.Paragraph("💬 I’d love to hear your feedback to help improve this library!") + md.LineBreak(1))

	sb.WriteString(md.Paragraph(ufmt.Sprintf(
		"🌐 You can check out a demo of this package in action at %s.",
		md.Link("gno.land/r/sunspirit/md", "/r/sunspirit/md"),
	)) + md.LineBreak(1))
	sb.WriteString(md.HorizontalRule())

	return sb.String()
}
