package markdown

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

// The validator inspects a destination on its way to the browser, so it
// has to see the resolved value: goldmark turns `&#x64;ata:` into
// `data:` when it emits the src, and a validator fed the raw bytes would
// pass a URL it is there to reject.
func TestImgValidatorSeesResolvedDestination(t *testing.T) {
	// AllowSvgDataImage is the policy gnoweb wires in (render_config.go), not
	// a copy of it — so reverting the policy fails this test.
	render := func(t *testing.T, src string) string {
		t.Helper()
		gnourl, err := weburl.Parse("https://gno.land/r/test")
		require.NoError(t, err)
		m := goldmark.New()
		NewGnoExtension(WithImageValidator(AllowSvgDataImage)).Extend(m)
		ctx := parser.WithContext(NewGnoParserContext(GnoContext{GnoURL: gnourl}))
		var out bytes.Buffer
		require.NoError(t, m.Convert([]byte(src), &out, ctx))
		return out.String()
	}

	// Baseline: the policy rejects a raster data: URI outright.
	require.Contains(t, render(t, `![i](data:image/png;base64,iVBORw0KGgo=)`), `<img src=""`)

	// The same URI behind an entity reference must be rejected too.
	//
	// The capitalised spellings matter independently: goldmark's
	// IsDangerousURL matches `data:image/png` case-insensitively and permits
	// it, so if this policy compared case-sensitively the raster URI would
	// reach the browser with no entity encoding needed at all.
	for _, src := range []string{
		`![i](&#x64;ata:image/png;base64,iVBORw0KGgo=)`,
		`![i](&#100;ata:image/gif;base64,R0lGODlh)`,
		`![i](d&#x61;ta:image/jpeg;base64,/9j/4AAQ)`,
		`![i](datA:image/png;base64,iVBORw0KGgo=)`,
		`![i](DATA:image/png;base64,iVBORw0KGgo=)`,
		`![i](Data:imAge/gif;base64,R0lGODlh)`,
		`![i](d&#X61;tA:imAge/png;base64,iVBORw0KGgo=)`,
	} {
		t.Run(src, func(t *testing.T) {
			require.Contains(t, render(t, src), `<img src=""`)
		})
	}

	// Allowed destinations still render — including a capitalised svg data
	// URI, so the case-insensitive match does not over-reject.
	require.Contains(t, render(t, `![i](https://ok.example/a.png)`), `src="https://ok.example/a.png"`)
	require.Contains(t, render(t, `![i](&#x64;ata:image/svg+xml;base64,PHN2Zz48L3N2Zz4=)`),
		`src="data:image/svg+xml;base64,PHN2Zz48L3N2Zz4="`)
	require.Contains(t, render(t, `![i](DATA:image/SVG+xml;base64,PHN2Zz48L3N2Zz4=)`),
		`src="DATA:image/SVG+xml;base64,PHN2Zz48L3N2Zz4="`)
}

// A leading C0 control or space is not part of the URL a browser parses, but
// it does shift the prefix: without a trim, `\tdata:image/png;…` shows the
// policy no "data:" and is waved through. URLEscape happens to neutralize the
// result today (the byte becomes %09, so the browser reads a relative path),
// so this is the policy's own boundary, not the last line of defense — the
// policy still has to hold on its own.
func TestAllowSvgDataImageTrimsLeadingControls(t *testing.T) {
	for _, uri := range []string{
		"\tdata:image/png;base64,iVBORw0KGgo=",
		" data:image/png;base64,iVBORw0KGgo=",
		"\ndata:text/html,<script>alert(1)</script>",
		"\x00data:image/gif;base64,R0lGODlh",
		"\r\n\t data:image/png;base64,iVBORw0KGgo=",
		"\tDATA:imAge/PNG;base64,iVBORw0KGgo=",
	} {
		t.Run(uri, func(t *testing.T) {
			require.False(t, AllowSvgDataImage(uri))
		})
	}

	// Still no over-rejection: a trimmed svg data URI, and anything that is
	// not a data: URI at all, stay allowed.
	require.True(t, AllowSvgDataImage("\tdata:image/svg+xml;base64,PHN2Zz48L3N2Zz4="))
	require.True(t, AllowSvgDataImage("\thttps://ok.example/a.png"))
	require.True(t, AllowSvgDataImage("/r/x:img.png"))
}
