package markdown

import (
	"bytes"
	"strconv"
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

	// GFM splits cells on '|' before inline parsing; &#124; is the escape
	// that survives (the label is raw text, so `\|` would keep its backslash).
	out = renderMarkdown(t, "| A |\n| --- |\n| <gno-button href=\"/r/test\" label=\"a&#124;b\" /> |\n")
	require.Contains(t, out, `<td><a href="/r/test" class="gno-button">a|b</a></td>`)
}

// The href is decoded once, by the tokenizer, as HTML decodes attributes; the
// escaper makes the link pipeline's resolveDestination give it back unchanged.
func TestButtonDestinationDecodedOnce(t *testing.T) {
	for _, href := range []string{"/r/a?x=&lt;b&gt;", "&amp;", "&#106;avascript:", `/r/a\_b`, `\\`, "a&b", "&", `\`} {
		got := resolveDestination([]byte(buttonDestEscaper.Replace(href)))
		require.Equal(t, href, string(got))
	}
}

// A parse attempt reads at most MaxButtonTagLen bytes, so a line of
// unterminated tags is linear (it was quadratic: each attempt tokenized to
// the end of the line). A tag whose `/>` lies past the window is not claimed.
func TestParseButtonTagBoundedWindow(t *testing.T) {
	tag := func(pad int) []byte {
		return []byte(`<gno-button href="/r/x" label="` + strings.Repeat("a", pad) + `" />`)
	}
	base := len(tag(0))

	_, n, ok := parseButtonTag(tag(MaxButtonTagLen - base))
	require.True(t, ok)
	require.Equal(t, MaxButtonTagLen, n)

	_, _, ok = parseButtonTag(tag(MaxButtonTagLen - base + 1))
	require.False(t, ok)
}

// Every '<' in a document reaches the button parsers; one that is not a
// button tag must cost no allocation.
func TestParseButtonTagNoAllocOnMiss(t *testing.T) {
	for _, in := range []string{
		"<div>", "<gno-columns>", "<gno-buttons />", "<",
		strings.Repeat(`<gno-button href="/r/x" `, 50_000), // unterminated, huge line
	} {
		b := []byte(in)
		allocs := testing.AllocsPerRun(100, func() { parseButtonTag(b) })
		require.Zero(t, allocs, in)
	}
}

// benchConvert renders src with the production extension on every iteration.
func benchConvert(b *testing.B, src []byte) {
	b.Helper()
	gnourl, _ := weburl.Parse("https://gno.land/r/test")
	m := goldmark.New(goldmark.WithExtensions(NewGnoExtension()))
	b.ReportAllocs()
	for b.Loop() {
		var buf bytes.Buffer
		ctx := parser.WithContext(NewGnoParserContext(GnoContext{GnoURL: gnourl}))
		if err := m.Convert(src, &buf, ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkButton compares a page of buttons with the same page written as
// plain markdown links, which go through the same link pipeline.
func BenchmarkButton(b *testing.B) {
	for name, item := range map[string]string{
		"button": `<gno-button href="/r/gnoland/blog" label="Read the blog" variant="outline" />`,
		"link":   `[Read the blog](/r/gnoland/blog)`,
	} {
		src := []byte(strings.Repeat("Text "+item+" and <span>html</span>.\n\n", 200))
		b.Run(name, func(b *testing.B) {
			benchConvert(b, src)
		})
	}
}

// BenchmarkButtonHostileLine renders one line of n unterminated tags. From
// n=1000 to n=4000, ns/op must grow about 4x (linear); it grew about 16x
// (quadratic) before the tag window. The "slash" shape puts `/>` inside each
// value so every attempt reaches the tokenizer.
func BenchmarkButtonHostileLine(b *testing.B) {
	for _, shape := range []struct{ name, item string }{
		{"unterminated", `<gno-button href="/r/x" `},
		{"slash", `<gno-button href="/>" `},
	} {
		for _, n := range []int{1000, 4000} {
			src := []byte("a " + strings.Repeat(shape.item, n))
			b.Run(shape.name+"/n="+strconv.Itoa(n), func(b *testing.B) {
				benchConvert(b, src)
			})
		}
	}
}
