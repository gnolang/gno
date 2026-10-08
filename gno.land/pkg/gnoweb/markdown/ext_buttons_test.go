package markdown

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

// The golden runner loads no GFM extension; renderMarkdown does.
func TestButtonInTable(t *testing.T) {
	out := renderMarkdown(t, "| Action | Button |\n| --- | --- |\n| Read | <gno-button href=\"/r/test\" label=\"Open\" variant=\"outline\" /> |\n")
	require.Contains(t, out, `<td><a href="/r/test" class="gno-button gno-button-outline">Open</a></td>`)
}

// Every '<' in a document reaches the button parsers; one that is not a
// button tag must cost no allocation.
func TestParseButtonTagNoAllocOnMiss(t *testing.T) {
	for _, in := range []string{"<div>", "<gno-columns>", "<gno-buttons />", "<"} {
		b := []byte(in)
		allocs := testing.AllocsPerRun(100, func() { parseButtonTag(b) })
		require.Zero(t, allocs, in)
	}
}

// BenchmarkButton compares a page of buttons with the same page written as
// plain markdown links, which go through the same link pipeline.
func BenchmarkButton(b *testing.B) {
	gnourl, _ := weburl.Parse("https://gno.land/r/test")
	m := goldmark.New(goldmark.WithExtensions(NewGnoExtension()))

	for name, item := range map[string]string{
		"button": `<gno-button href="/r/gnoland/blog" label="Read the blog" variant="outline" />`,
		"link":   `[Read the blog](/r/gnoland/blog)`,
	} {
		src := []byte(strings.Repeat("Text "+item+" and <span>html</span>.\n\n", 200))
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				var buf bytes.Buffer
				ctx := parser.WithContext(NewGnoParserContext(GnoContext{GnoURL: gnourl}))
				if err := m.Convert(src, &buf, ctx); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
