package markdown

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

// BenchmarkPanel compares a document with and without panel wrappers, to
// keep the per-line cost of panelParser.Continue visible.
func BenchmarkPanel(b *testing.B) {
	body := strings.Repeat("Some **text** with a [link](/r/demo).\n\n- item\n- item\n\n<b>inline html</b>\n\n", 20)
	docs := map[string]string{
		"plain": strings.Repeat(body, 10),
		"panel": strings.Repeat("<gno-panel>\n"+body+"</gno-panel>\n\n", 10),
	}
	for name, doc := range docs {
		src := []byte(doc)
		b.Run(name, func(b *testing.B) {
			m := goldmark.New()
			NewGnoExtension().Extend(m)
			var buf bytes.Buffer
			b.ReportAllocs()
			for b.Loop() {
				buf.Reset()
				ctx := parser.WithContext(NewGnoParserContext(GnoContext{}))
				if err := m.Convert(src, &buf, ctx); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
