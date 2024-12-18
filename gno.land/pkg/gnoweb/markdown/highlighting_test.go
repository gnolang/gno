// This file was copied from https://github.com/yuin/goldmark-highlighting

package markdown

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/testutil"
	"github.com/yuin/goldmark/util"
)

func TestHighlighting(t *testing.T) {
	var css bytes.Buffer
	markdown := goldmark.New(
		goldmark.WithExtensions(
			NewHighlighting(
				WithStyle("monokai"),
				WithCSSWriter(&css),
				WithFormatOptions(
					chromahtml.WithClasses(true),
					chromahtml.WithLineNumbers(false),
				),
				WithWrapperRenderer(func(w util.BufWriter, c CodeBlockContext, entering bool) {
					_, ok := c.Language()
					if entering {
						if !ok {
							w.WriteString("<pre><code>")
							return
						}
						w.WriteString(`<div class="highlight">`)
					} else {
						if !ok {
							w.WriteString("</pre></code>")
							return
						}
						w.WriteString(`</div>`)
					}
				}),
				WithCodeBlockOptions(func(c CodeBlockContext) []chromahtml.Option {
					if language, ok := c.Language(); ok {
						// Turn on line numbers for Go only.
						if string(language) == "go" {
							return []chromahtml.Option{
								chromahtml.WithLineNumbers(true),
							}
						}
					}
					return nil
				}),
			),
		),
	)
	var buffer bytes.Buffer
	if err := markdown.Convert([]byte(`
Title
=======
`+"``` go\n"+`func main() {
    fmt.Println("ok")
}
`+"```"+`
`), &buffer); err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(buffer.String()) != strings.TrimSpace(`
<h1>Title</h1>
<div class="highlight"><pre class="chroma"><code><span class="line"><span class="ln">1</span><span class="cl"><span class="kd">func</span> <span class="nf">main</span><span class="p">()</span> <span class="p">{</span>
</span></span><span class="line"><span class="ln">2</span><span class="cl">    <span class="nx">fmt</span><span class="p">.</span><span class="nf">Println</span><span class="p">(</span><span class="s">&#34;ok&#34;</span><span class="p">)</span>
</span></span><span class="line"><span class="ln">3</span><span class="cl"><span class="p">}</span>
</span></span></code></pre></div>
`) {
		t.Errorf("failed to render HTML\n%s", buffer.String())
	}

	expected := strings.TrimSpace(`/* Background */ .bg { color: #f8f8f2; background-color: #272822; }
/* PreWrapper */ .chroma { color: #f8f8f2; background-color: #272822; }
/* LineNumbers targeted by URL anchor */ .chroma .ln:target { color: #f8f8f2; background-color: #3c3d38 }
/* LineNumbersTable targeted by URL anchor */ .chroma .lnt:target { color: #f8f8f2; background-color: #3c3d38 }
/* Error */ .chroma .err { color: #960050; background-color: #1e0010 }
/* LineLink */ .chroma .lnlinks { outline: none; text-decoration: none; color: inherit }
/* LineTableTD */ .chroma .lntd { vertical-align: top; padding: 0; margin: 0; border: 0; }
/* LineTable */ .chroma .lntable { border-spacing: 0; padding: 0; margin: 0; border: 0; }
/* LineHighlight */ .chroma .hl { background-color: #3c3d38 }
/* LineNumbersTable */ .chroma .lnt { white-space: pre; -webkit-user-select: none; user-select: none; margin-right: 0.4em; padding: 0 0.4em 0 0.4em;color: #7f7f7f }
/* LineNumbers */ .chroma .ln { white-space: pre; -webkit-user-select: none; user-select: none; margin-right: 0.4em; padding: 0 0.4em 0 0.4em;color: #7f7f7f }
/* Line */ .chroma .line { display: flex; }
/* Keyword */ .chroma .k { color: #66d9ef }
/* KeywordConstant */ .chroma .kc { color: #66d9ef }
/* KeywordDeclaration */ .chroma .kd { color: #66d9ef }
/* KeywordNamespace */ .chroma .kn { color: #f92672 }
/* KeywordPseudo */ .chroma .kp { color: #66d9ef }
/* KeywordReserved */ .chroma .kr { color: #66d9ef }
/* KeywordType */ .chroma .kt { color: #66d9ef }
/* NameAttribute */ .chroma .na { color: #a6e22e }
/* NameClass */ .chroma .nc { color: #a6e22e }
/* NameConstant */ .chroma .no { color: #66d9ef }
/* NameDecorator */ .chroma .nd { color: #a6e22e }
/* NameException */ .chroma .ne { color: #a6e22e }
/* NameFunction */ .chroma .nf { color: #a6e22e }
/* NameOther */ .chroma .nx { color: #a6e22e }
/* NameTag */ .chroma .nt { color: #f92672 }
/* Literal */ .chroma .l { color: #ae81ff }
/* LiteralDate */ .chroma .ld { color: #e6db74 }
/* LiteralString */ .chroma .s { color: #e6db74 }
/* LiteralStringAffix */ .chroma .sa { color: #e6db74 }
/* LiteralStringBacktick */ .chroma .sb { color: #e6db74 }
/* LiteralStringChar */ .chroma .sc { color: #e6db74 }
/* LiteralStringDelimiter */ .chroma .dl { color: #e6db74 }
/* LiteralStringDoc */ .chroma .sd { color: #e6db74 }
/* LiteralStringDouble */ .chroma .s2 { color: #e6db74 }
/* LiteralStringEscape */ .chroma .se { color: #ae81ff }
/* LiteralStringHeredoc */ .chroma .sh { color: #e6db74 }
/* LiteralStringInterpol */ .chroma .si { color: #e6db74 }
/* LiteralStringOther */ .chroma .sx { color: #e6db74 }
/* LiteralStringRegex */ .chroma .sr { color: #e6db74 }
/* LiteralStringSingle */ .chroma .s1 { color: #e6db74 }
/* LiteralStringSymbol */ .chroma .ss { color: #e6db74 }
/* LiteralNumber */ .chroma .m { color: #ae81ff }
/* LiteralNumberBin */ .chroma .mb { color: #ae81ff }
/* LiteralNumberFloat */ .chroma .mf { color: #ae81ff }
/* LiteralNumberHex */ .chroma .mh { color: #ae81ff }
/* LiteralNumberInteger */ .chroma .mi { color: #ae81ff }
/* LiteralNumberIntegerLong */ .chroma .il { color: #ae81ff }
/* LiteralNumberOct */ .chroma .mo { color: #ae81ff }
/* Operator */ .chroma .o { color: #f92672 }
/* OperatorWord */ .chroma .ow { color: #f92672 }
/* Comment */ .chroma .c { color: #75715e }
/* CommentHashbang */ .chroma .ch { color: #75715e }
/* CommentMultiline */ .chroma .cm { color: #75715e }
/* CommentSingle */ .chroma .c1 { color: #75715e }
/* CommentSpecial */ .chroma .cs { color: #75715e }
/* CommentPreproc */ .chroma .cp { color: #75715e }
/* CommentPreprocFile */ .chroma .cpf { color: #75715e }
/* GenericDeleted */ .chroma .gd { color: #f92672 }
/* GenericEmph */ .chroma .ge { font-style: italic }
/* GenericInserted */ .chroma .gi { color: #a6e22e }
/* GenericStrong */ .chroma .gs { font-weight: bold }
/* GenericSubheading */ .chroma .gu { color: #75715e }`)

	gotten := strings.TrimSpace(css.String())

	if expected != gotten {
		diff := testutil.DiffPretty([]byte(expected), []byte(gotten))
		t.Errorf("incorrect CSS.\n%s", string(diff))
	}
}

func TestHighlighting2(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			Highlighting,
		),
	)
	var buffer bytes.Buffer
	if err := markdown.Convert([]byte(`
Title
=======
`+"```"+`
func main() {
    fmt.Println("ok")
}
`+"```"+`
`), &buffer); err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(buffer.String()) != strings.TrimSpace(`
<h1>Title</h1>
<pre><code>func main() {
    fmt.Println(&quot;ok&quot;)
}
</code></pre>
`) {
		t.Error("failed to render HTML")
	}
}

func TestHighlighting3(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			Highlighting,
		),
	)
	var buffer bytes.Buffer
	if err := markdown.Convert([]byte(`
Title
=======

`+"```"+`cpp {hl_lines=[1,2]}
#include <iostream>
int main() {
    std::cout<< "hello" << std::endl;
}
`+"```"+`
`), &buffer); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buffer.String()) != strings.TrimSpace(`
<h1>Title</h1>
<pre style="background-color:#fff;display:grid;"><code><span style="display:flex; background-color:#e5e5e5"><span><span style="color:#999;font-weight:bold;font-style:italic">#include</span> <span style="color:#999;font-weight:bold;font-style:italic">&lt;iostream&gt;</span><span style="color:#999;font-weight:bold;font-style:italic">
</span></span></span><span style="display:flex; background-color:#e5e5e5"><span><span style="color:#999;font-weight:bold;font-style:italic"></span><span style="color:#458;font-weight:bold">int</span> <span style="color:#900;font-weight:bold">main</span>() {
</span></span><span style="display:flex;"><span>    std<span style="color:#000;font-weight:bold">::</span>cout<span style="color:#000;font-weight:bold">&lt;&lt;</span> <span style="color:#d14">&#34;hello&#34;</span> <span style="color:#000;font-weight:bold">&lt;&lt;</span> std<span style="color:#000;font-weight:bold">::</span>endl;
</span></span><span style="display:flex;"><span>}
</span></span></code></pre>
`) {
		t.Errorf("failed to render HTML:\n%s", buffer.String())
	}
}

func TestHighlightingCustom(t *testing.T) {
	custom := chroma.MustNewStyle("custom", chroma.StyleEntries{
		chroma.Background:           "#cccccc bg:#1d1d1d",
		chroma.Comment:              "#999999",
		chroma.CommentSpecial:       "#cd0000",
		chroma.Keyword:              "#cc99cd",
		chroma.KeywordDeclaration:   "#cc99cd",
		chroma.KeywordNamespace:     "#cc99cd",
		chroma.KeywordType:          "#cc99cd",
		chroma.Operator:             "#67cdcc",
		chroma.OperatorWord:         "#cdcd00",
		chroma.NameClass:            "#f08d49",
		chroma.NameBuiltin:          "#f08d49",
		chroma.NameFunction:         "#f08d49",
		chroma.NameException:        "bold #666699",
		chroma.NameVariable:         "#00cdcd",
		chroma.LiteralString:        "#7ec699",
		chroma.LiteralNumber:        "#f08d49",
		chroma.LiteralStringBoolean: "#f08d49",
		chroma.GenericHeading:       "bold #000080",
		chroma.GenericSubheading:    "bold #800080",
		chroma.GenericDeleted:       "#e2777a",
		chroma.GenericInserted:      "#cc99cd",
		chroma.GenericError:         "#e2777a",
		chroma.GenericEmph:          "italic",
		chroma.GenericStrong:        "bold",
		chroma.GenericPrompt:        "bold #000080",
		chroma.GenericOutput:        "#888",
		chroma.GenericTraceback:     "#04D",
		chroma.GenericUnderline:     "underline",
		chroma.Error:                "border:#e2777a",
	})

	var css bytes.Buffer
	markdown := goldmark.New(
		goldmark.WithExtensions(
			NewHighlighting(
				WithStyle("monokai"), // to make sure it is overrided even if present
				WithCustomStyle(custom),
				WithCSSWriter(&css),
				WithFormatOptions(
					chromahtml.WithClasses(true),
					chromahtml.WithLineNumbers(false),
				),
				WithWrapperRenderer(func(w util.BufWriter, c CodeBlockContext, entering bool) {
					_, ok := c.Language()
					if entering {
						if !ok {
							w.WriteString("<pre><code>")
							return
						}
						w.WriteString(`<div class="highlight">`)
					} else {
						if !ok {
							w.WriteString("</pre></code>")
							return
						}
						w.WriteString(`</div>`)
					}
				}),
				WithCodeBlockOptions(func(c CodeBlockContext) []chromahtml.Option {
					if language, ok := c.Language(); ok {
						// Turn on line numbers for Go only.
						if string(language) == "go" {
							return []chromahtml.Option{
								chromahtml.WithLineNumbers(true),
							}
						}
					}
					return nil
				}),
			),
		),
	)
	var buffer bytes.Buffer
	if err := markdown.Convert([]byte(`
Title
=======
`+"``` go\n"+`func main() {
    fmt.Println("ok")
}
`+"```"+`
`), &buffer); err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(buffer.String()) != strings.TrimSpace(`
<h1>Title</h1>
<div class="highlight"><pre class="chroma"><code><span class="line"><span class="ln">1</span><span class="cl"><span class="kd">func</span> <span class="nf">main</span><span class="p">()</span> <span class="p">{</span>
</span></span><span class="line"><span class="ln">2</span><span class="cl">    <span class="nx">fmt</span><span class="p">.</span><span class="nf">Println</span><span class="p">(</span><span class="s">&#34;ok&#34;</span><span class="p">)</span>
</span></span><span class="line"><span class="ln">3</span><span class="cl"><span class="p">}</span>
</span></span></code></pre></div>
`) {
		t.Error("failed to render HTML", buffer.String())
	}

	expected := strings.TrimSpace(`/* Background */ .bg { color: #cccccc; background-color: #1d1d1d; }
/* PreWrapper */ .chroma { color: #cccccc; background-color: #1d1d1d; }
/* LineNumbers targeted by URL anchor */ .chroma .ln:target { color: #cccccc; background-color: #333333 }
/* LineNumbersTable targeted by URL anchor */ .chroma .lnt:target { color: #cccccc; background-color: #333333 }
/* Error */ .chroma .err {  }
/* LineLink */ .chroma .lnlinks { outline: none; text-decoration: none; color: inherit }
/* LineTableTD */ .chroma .lntd { vertical-align: top; padding: 0; margin: 0; border: 0; }
/* LineTable */ .chroma .lntable { border-spacing: 0; padding: 0; margin: 0; border: 0; }
/* LineHighlight */ .chroma .hl { background-color: #333333 }
/* LineNumbersTable */ .chroma .lnt { white-space: pre; -webkit-user-select: none; user-select: none; margin-right: 0.4em; padding: 0 0.4em 0 0.4em;color: #666666 }
/* LineNumbers */ .chroma .ln { white-space: pre; -webkit-user-select: none; user-select: none; margin-right: 0.4em; padding: 0 0.4em 0 0.4em;color: #666666 }
/* Line */ .chroma .line { display: flex; }
/* Keyword */ .chroma .k { color: #cc99cd }
/* KeywordConstant */ .chroma .kc { color: #cc99cd }
/* KeywordDeclaration */ .chroma .kd { color: #cc99cd }
/* KeywordNamespace */ .chroma .kn { color: #cc99cd }
/* KeywordPseudo */ .chroma .kp { color: #cc99cd }
/* KeywordReserved */ .chroma .kr { color: #cc99cd }
/* KeywordType */ .chroma .kt { color: #cc99cd }
/* NameBuiltin */ .chroma .nb { color: #f08d49 }
/* NameClass */ .chroma .nc { color: #f08d49 }
/* NameException */ .chroma .ne { color: #666699; font-weight: bold }
/* NameFunction */ .chroma .nf { color: #f08d49 }
/* NameVariable */ .chroma .nv { color: #00cdcd }
/* LiteralString */ .chroma .s { color: #7ec699 }
/* LiteralStringAffix */ .chroma .sa { color: #7ec699 }
/* LiteralStringBacktick */ .chroma .sb { color: #7ec699 }
/* LiteralStringChar */ .chroma .sc { color: #7ec699 }
/* LiteralStringDelimiter */ .chroma .dl { color: #7ec699 }
/* LiteralStringDoc */ .chroma .sd { color: #7ec699 }
/* LiteralStringDouble */ .chroma .s2 { color: #7ec699 }
/* LiteralStringEscape */ .chroma .se { color: #7ec699 }
/* LiteralStringHeredoc */ .chroma .sh { color: #7ec699 }
/* LiteralStringInterpol */ .chroma .si { color: #7ec699 }
/* LiteralStringOther */ .chroma .sx { color: #7ec699 }
/* LiteralStringRegex */ .chroma .sr { color: #7ec699 }
/* LiteralStringSingle */ .chroma .s1 { color: #7ec699 }
/* LiteralStringSymbol */ .chroma .ss { color: #7ec699 }
/* LiteralNumber */ .chroma .m { color: #f08d49 }
/* LiteralNumberBin */ .chroma .mb { color: #f08d49 }
/* LiteralNumberFloat */ .chroma .mf { color: #f08d49 }
/* LiteralNumberHex */ .chroma .mh { color: #f08d49 }
/* LiteralNumberInteger */ .chroma .mi { color: #f08d49 }
/* LiteralNumberIntegerLong */ .chroma .il { color: #f08d49 }
/* LiteralNumberOct */ .chroma .mo { color: #f08d49 }
/* Operator */ .chroma .o { color: #67cdcc }
/* OperatorWord */ .chroma .ow { color: #cdcd00 }
/* Comment */ .chroma .c { color: #999999 }
/* CommentHashbang */ .chroma .ch { color: #999999 }
/* CommentMultiline */ .chroma .cm { color: #999999 }
/* CommentSingle */ .chroma .c1 { color: #999999 }
/* CommentSpecial */ .chroma .cs { color: #cd0000 }
/* CommentPreproc */ .chroma .cp { color: #999999 }
/* CommentPreprocFile */ .chroma .cpf { color: #999999 }
/* GenericDeleted */ .chroma .gd { color: #e2777a }
/* GenericEmph */ .chroma .ge { font-style: italic }
/* GenericError */ .chroma .gr { color: #e2777a }
/* GenericHeading */ .chroma .gh { color: #000080; font-weight: bold }
/* GenericInserted */ .chroma .gi { color: #cc99cd }
/* GenericOutput */ .chroma .go { color: #888888 }
/* GenericPrompt */ .chroma .gp { color: #000080; font-weight: bold }
/* GenericStrong */ .chroma .gs { font-weight: bold }
/* GenericSubheading */ .chroma .gu { color: #800080; font-weight: bold }
/* GenericTraceback */ .chroma .gt { color: #0044dd }
/* GenericUnderline */ .chroma .gl { text-decoration: underline }`)

	gotten := strings.TrimSpace(css.String())

	if expected != gotten {
		diff := testutil.DiffPretty([]byte(expected), []byte(gotten))
		t.Errorf("incorrect CSS.\n%s", string(diff))
	}
}

func TestHighlightingHlLines(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			NewHighlighting(
				WithFormatOptions(
					chromahtml.WithClasses(true),
				),
			),
		),
	)

	for i, test := range []struct {
		attributes string
		expect     []int
	}{
		{`hl_lines=["2"]`, []int{2}},
		{`hl_lines=["2-3",5],linenostart=5`, []int{2, 3, 5}},
		{`hl_lines=["2-3"]`, []int{2, 3}},
		{`hl_lines=["2-3",5],linenostart="5"`, []int{2, 3}}, // linenostart must be a number. string values are ignored
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			var buffer bytes.Buffer
			codeBlock := fmt.Sprintf(`bash {%s}
LINE1
LINE2
LINE3
LINE4
LINE5
LINE6
LINE7
LINE8
`, test.attributes)

			if err := markdown.Convert([]byte(`
`+"```"+codeBlock+"```"+`
`), &buffer); err != nil {
				t.Fatal(err)
			}

			for _, line := range test.expect {
				expectStr := fmt.Sprintf("<span class=\"line hl\"><span class=\"cl\">LINE%d\n</span></span>", line)
				if !strings.Contains(buffer.String(), expectStr) {
					t.Fatal("got\n", buffer.String(), "\nexpected\n", expectStr)
				}
			}
		})
	}
}

type nopPreWrapper struct{}

// Start is called to write a start <pre> element.
func (nopPreWrapper) Start(code bool, styleAttr string) string { return "" }

// End is called to write the end </pre> element.
func (nopPreWrapper) End(code bool) string { return "" }

func TestHighlightingLinenos(t *testing.T) {
	outputLineNumbersInTable := `<div class="chroma">
<table class="lntable"><tr><td class="lntd">
<span class="lnt">1
</span></td>
<td class="lntd">
<span class="line"><span class="cl">LINE1
</span></span></td></tr></table>
</div>`

	for i, test := range []struct {
		attributes         string
		lineNumbers        bool
		lineNumbersInTable bool
		expect             string
	}{
		{`linenos=true`, false, false, `<span class="line"><span class="ln">1</span><span class="cl">LINE1
</span></span>`},
		{`linenos=false`, false, false, `<span class="line"><span class="cl">LINE1
</span></span>`},
		{``, true, false, `<span class="line"><span class="ln">1</span><span class="cl">LINE1
</span></span>`},
		{``, true, true, outputLineNumbersInTable},
		{`linenos=inline`, true, true, `<span class="line"><span class="ln">1</span><span class="cl">LINE1
</span></span>`},
		{`linenos=foo`, false, false, `<span class="line"><span class="ln">1</span><span class="cl">LINE1
</span></span>`},
		{`linenos=table`, false, false, outputLineNumbersInTable},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			markdown := goldmark.New(
				goldmark.WithExtensions(
					NewHighlighting(
						WithFormatOptions(
							chromahtml.WithLineNumbers(test.lineNumbers),
							chromahtml.LineNumbersInTable(test.lineNumbersInTable),
							chromahtml.WithPreWrapper(nopPreWrapper{}),
							chromahtml.WithClasses(true),
						),
					),
				),
			)

			var buffer bytes.Buffer
			codeBlock := fmt.Sprintf(`bash {%s}
LINE1
`, test.attributes)

			content := "```" + codeBlock + "```"

			if err := markdown.Convert([]byte(content), &buffer); err != nil {
				t.Fatal(err)
			}

			s := strings.TrimSpace(buffer.String())

			if s != test.expect {
				t.Fatal("got\n", s, "\nexpected\n", test.expect)
			}
		})
	}
}

func TestHighlightingGuessLanguage(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			NewHighlighting(
				WithGuessLanguage(true),
				WithFormatOptions(
					chromahtml.WithClasses(true),
					chromahtml.WithLineNumbers(true),
				),
			),
		),
	)
	var buffer bytes.Buffer
	if err := markdown.Convert([]byte("```"+`
LINE
`+"```"), &buffer); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buffer.String()) != strings.TrimSpace(`
<pre class="chroma"><code><span class="line"><span class="ln">1</span><span class="cl">LINE
</span></span></code></pre>
`) {
		t.Errorf("render mismatch, got\n%s", buffer.String())
	}
}

func TestCoalesceNeeded(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			NewHighlighting(
				// WithGuessLanguage(true),
				WithFormatOptions(
					chromahtml.WithClasses(true),
					chromahtml.WithLineNumbers(true),
				),
			),
		),
	)
	var buffer bytes.Buffer
	if err := markdown.Convert([]byte("```http"+`
GET /foo HTTP/1.1
Content-Type: application/json
User-Agent: foo

{
  "hello": "world"
}
`+"```"), &buffer); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buffer.String()) != strings.TrimSpace(`
<pre class="chroma"><code><span class="line"><span class="ln">1</span><span class="cl"><span class="nf">GET</span> <span class="nn">/foo</span> <span class="kr">HTTP</span><span class="o">/</span><span class="m">1.1</span>
</span></span><span class="line"><span class="ln">2</span><span class="cl"><span class="n">Content-Type</span><span class="o">:</span> <span class="l">application/json</span>
</span></span><span class="line"><span class="ln">3</span><span class="cl"><span class="n">User-Agent</span><span class="o">:</span> <span class="l">foo</span>
</span></span><span class="line"><span class="ln">4</span><span class="cl">
</span></span><span class="line"><span class="ln">5</span><span class="cl"><span class="p">{</span>
</span></span><span class="line"><span class="ln">6</span><span class="cl">  <span class="nt">&#34;hello&#34;</span><span class="p">:</span> <span class="s2">&#34;world&#34;</span>
</span></span><span class="line"><span class="ln">7</span><span class="cl"><span class="p">}</span>
</span></span></code></pre>
`) {
		t.Errorf("render mismatch, got\n%s", buffer.String())
	}
}
