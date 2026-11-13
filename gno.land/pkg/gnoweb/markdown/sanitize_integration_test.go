// Sanitize integration tests.
//
// Each .txtar case under golden/sanitize/ exercises one sanitize helper
// (declared via `// MARKDOWNFUNC: X` in the txtar comment), optionally
// substituting its output into a surrounding markdown template (declared
// via `// CONTEXT: <template>`) and rendering that with the same
// goldmark + gnoweb-extension chain used in production.
//
// Three sections:
//
//	-- input.md --     user-supplied content (the caller input)
//	-- output.md --    sanitize.X(input)  (skip-on-update generated)
//	-- output.html --  optional: rendered via goldmark+gnoweb when CONTEXT
//	                   is declared. Validators (no CONTEXT) omit this.
//
// Directives in the txtar comment (before the first `-- ... --`):
//
//	// MARKDOWNFUNC: <one of the sanitize.X names>
//	// CONTEXT: <markdown template containing exactly one %s — except
//	            CodeFence which uses two; `\n` in the template is
//	            interpreted as a real newline>
//	// ARGS: <Go literal — quoted string for BechString prefix,
//	          unquoted int for CodeFence minCount>
//
// Run with `-update-golden-tests` to regenerate output sections.
package markdown

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"golang.org/x/tools/txtar"
)

const sanitizeTestdataDir = "golden/sanitize"

// ----- gno VM driver -----
//
// Each case invokes the real gno helper at gno.land/p/nt/markdown/sanitize/v0
// through the gno VM (not a Go reimplementation). A fresh Machine is
// constructed per case from a shared base Store so cases stay isolated;
// the base Store loads the examples directory once.
//
// The driver gno file is loaded into each per-case Machine as a synthetic
// package that imports sanitize/v0 and exposes a single
// Run(fn, input, arg, arg2) dispatcher. The Go side then calls
// m.Eval(`Run("InlineText", "input", "", "")`) and pops the string
// result. arg2 is used only by the helpers that take three string slots
// (currently LinkReferenceDefinition: label + url + title).

const driverPkgPath = "gno.land/p/nt/markdown/sanitize/v0/sanitize_test_driver"

const driverSrc = `package sanitize_test_driver

import (
	"strconv"

	"gno.land/p/nt/markdown/sanitize/v0"
)

func Run(fn, input, arg, arg2 string) string {
	switch fn {
	case "StripBidiAndZeroWidth":
		return sanitize.StripBidiAndZeroWidth(input)
	case "NormalizeBreaks":
		return sanitize.NormalizeBreaks(input)
	case "InlineText":
		return sanitize.InlineText(input)
	case "Block":
		return sanitize.Block(input)
	case "BlockRich":
		return sanitize.BlockRich(input)
	case "LinkTitle":
		return sanitize.LinkTitle(input)
	case "TableCell":
		return sanitize.TableCell(input)
	case "HTMLEscape":
		return sanitize.HTMLEscape(input)
	case "URL":
		return sanitize.URL(input)
	case "ImageURL":
		return sanitize.ImageURL(input)
	case "UserName":
		return sanitize.UserName(input)
	case "BechString":
		return sanitize.BechString(input, arg)
	case "FootnoteLabel":
		return sanitize.FootnoteLabel(input)
	case "LanguageName":
		return sanitize.LanguageName(input)
	case "NestedPrefix":
		return sanitize.NestedPrefix(input)
	case "CodeFence":
		n, _ := strconv.Atoi(arg)
		return sanitize.CodeFence(input, n)
	case "InlineCode":
		return sanitize.InlineCode(input)
	case "CodeBlock":
		return sanitize.CodeBlock(input)
	case "LanguageCodeBlock":
		return sanitize.LanguageCodeBlock(arg, input)
	case "Blockquote":
		return sanitize.Blockquote(input)
	case "BlockquoteRich":
		return sanitize.BlockquoteRich(input)
	case "FootnoteDefinition":
		// arg = realm-provided footnote name; input = user body text.
		return sanitize.FootnoteDefinition(arg, input)
	case "LinkReferenceDefinition":
		// input = realm-provided label; arg = url; arg2 = title.
		return sanitize.LinkReferenceDefinition(input, arg, arg2)
	}
	panic("unknown sanitize function: " + fn)
}
`

// The driver package (which imports sanitize/v0 and its stdlib chain) is
// preprocessed exactly once and reused across every case. Preprocessing the
// import chain dominates runtime — re-running it per case made this the
// slowest package in the gno.land CI job — and the sanitize helpers are pure
// string→string functions with no mutable package state, so a single shared
// Machine is safe. Subtests here run sequentially (none call t.Parallel).
var (
	driverOnce    sync.Once
	driverMachine *gno.Machine
)

func sanitizeDriverMachine(t *testing.T) *gno.Machine {
	t.Helper()
	driverOnce.Do(func() {
		commit, gnoBase := test.TestStore(gnoenv.RootDir(), io.Discard, nil)
		tcw := commit.CacheWrap()
		tx := gnoBase.BeginTransaction(tcw, tcw, nil, nil)
		m := gno.NewMachineWithOptions(gno.MachineOptions{
			Store:         tx,
			Context:       test.Context("", driverPkgPath, nil),
			Output:        io.Discard,
			MaxAllocBytes: math.MaxInt64,
		})
		mpkg := &std.MemPackage{
			Type: gno.MPUserProd,
			Name: "sanitize_test_driver",
			Path: driverPkgPath,
			Files: []*std.MemFile{
				{Name: "gnomod.toml", Body: gno.GenGnoModLatest(driverPkgPath)},
				{Name: "driver.gno", Body: driverSrc},
			},
		}
		_, pv := m.RunMemPackage(mpkg, true)
		m.SetActivePackage(pv)
		driverMachine = m
	})
	return driverMachine
}

// callSanitizeGno Evals `Run(fn, input, arg, arg2)` against the shared driver
// Machine and returns the string result.
func callSanitizeGno(t *testing.T, fn, input, arg, arg2 string) string {
	t.Helper()
	m := sanitizeDriverMachine(t)

	// Build the call AST directly. gno.X parses a Go-shaped string with
	// chopBinary heuristics that misfire on quoted string literals, so
	// we construct CallExpr + BasicLitExpr arguments by hand.
	expr := gno.Call(gno.Nx("Run"), gno.Str(fn), gno.Str(input), gno.Str(arg), gno.Str(arg2))
	results := m.Eval(expr)
	require.Len(t, results, 1, "Run() must return exactly one value")
	return results[0].GetString()
}

// ----- directive parsing -----

type sanitizeCase struct {
	Func    string
	Context string // empty if no CONTEXT directive — validators
	Args    string // raw text after `// ARGS:` — empty if absent
	Escaped bool   // when true, decode Go-string escapes in input section
}

var (
	reMDFunc  = regexp.MustCompile(`(?m)^//\s*MARKDOWNFUNC:\s*(\S+)\s*$`)
	reContext = regexp.MustCompile(`(?m)^//\s*CONTEXT:\s*(.*)$`)
	reArgs    = regexp.MustCompile(`(?m)^//\s*ARGS:\s*(.*)$`)
	reEscaped = regexp.MustCompile(`(?m)^//\s*INPUT_ESCAPED\s*$`)
)

func parseDirectives(t *testing.T, comment []byte) sanitizeCase {
	t.Helper()
	c := sanitizeCase{}
	if m := reMDFunc.FindSubmatch(comment); m != nil {
		c.Func = string(m[1])
	}
	if m := reContext.FindSubmatch(comment); m != nil {
		c.Context = strings.ReplaceAll(string(m[1]), `\n`, "\n")
	}
	if m := reArgs.FindSubmatch(comment); m != nil {
		c.Args = strings.TrimSpace(string(m[1]))
	}
	if reEscaped.Match(comment) {
		c.Escaped = true
	}
	require.NotEmpty(t, c.Func, "missing // MARKDOWNFUNC: directive")
	return c
}

// decodeInput unescapes Go-string-style escapes (\r, \n, \t, \x00,
// \u202E, etc.) in the raw input bytes. Used when the case declares
// `// INPUT_ESCAPED` to author bytes that Write/editor tooling would
// otherwise normalize (CR, NUL, lone control bytes).
func decodeInput(t *testing.T, raw string) string {
	t.Helper()
	// Escape any literal double-quotes so the wrap-and-Unquote works.
	quoted := `"` + strings.ReplaceAll(raw, `"`, `\"`) + `"`
	out, err := strconv.Unquote(quoted)
	require.NoError(t, err, "INPUT_ESCAPED decode failed for %q", raw)
	return out
}

// ----- dispatch -----

// applySanitize calls the named sanitize function through the gno VM.
// sc.Args is passed through to the driver verbatim (raw text after
// `// ARGS:`); the driver dispatches on sc.Func and parses sc.Args as
// needed (strconv.Unquote for BechString prefix, strconv.Atoi for
// CodeFence minCount). This way the policy stays in one place — the
// gno-side driver — and the Go test code just shuttles strings.
func applySanitize(t *testing.T, sc sanitizeCase, input string) string {
	t.Helper()
	arg, arg2 := "", ""
	switch sc.Func {
	case "BechString":
		// BechString's ARGS is a Go string literal; unquote so the gno
		// helper receives the raw prefix.
		unq, err := strconv.Unquote(sc.Args)
		require.NoError(t, err, "BechString ARGS must be a quoted Go string literal, got %q", sc.Args)
		arg = unq
	case "LanguageCodeBlock":
		// Language tag is a Go string literal; unquote so the gno helper
		// receives the raw tag (which it then validates via LanguageName).
		unq, err := strconv.Unquote(sc.Args)
		require.NoError(t, err, "LanguageCodeBlock ARGS must be a quoted Go string literal, got %q", sc.Args)
		arg = unq
	case "FootnoteDefinition":
		// FootnoteDefinition's ARGS is the realm-provided footnote name
		// as a quoted Go string literal (input.md holds the body text).
		unq, err := strconv.Unquote(sc.Args)
		require.NoError(t, err, "FootnoteDefinition ARGS must be a quoted Go string literal, got %q", sc.Args)
		arg = unq
	case "LinkReferenceDefinition":
		// LinkReferenceDefinition's ARGS is two quoted Go string literals
		// separated by a comma: "url","title". The label comes from
		// input.md. The title may be the empty string "".
		url, title, ok := splitTwoQuoted(sc.Args)
		require.True(t, ok, "LinkReferenceDefinition ARGS must be two comma-separated quoted Go string literals, got %q", sc.Args)
		arg = url
		arg2 = title
	case "CodeFence":
		// CodeFence's ARGS is an integer; driver re-parses with
		// strconv.Atoi, so pass through as-is.
		arg = sc.Args
	}
	return callSanitizeGno(t, sc.Func, input, arg, arg2)
}

// splitTwoQuoted parses two comma-separated Go-quoted string literals,
// tolerating whitespace around the comma. Used for cases that need to
// thread two realm-provided slots through ARGS.
func splitTwoQuoted(s string) (a, b string, ok bool) {
	first, err := strconv.QuotedPrefix(s)
	if err != nil {
		return "", "", false
	}
	rest := strings.TrimLeft(s[len(first):], " \t")
	if !strings.HasPrefix(rest, ",") {
		return "", "", false
	}
	rest = strings.TrimLeft(rest[1:], " \t")
	second, err := strconv.QuotedPrefix(rest)
	if err != nil {
		return "", "", false
	}
	if strings.TrimSpace(rest[len(second):]) != "" {
		return "", "", false
	}
	au, err := strconv.Unquote(first)
	if err != nil {
		return "", "", false
	}
	bu, err := strconv.Unquote(second)
	if err != nil {
		return "", "", false
	}
	return au, bu, true
}

// substituteContext substitutes the sanitize output into the CONTEXT
// template. CodeFence uses the output twice (open + close fence).
func substituteContext(sc sanitizeCase, output string) string {
	count := strings.Count(sc.Context, "%s")
	args := make([]any, count)
	for i := range args {
		args[i] = output
	}
	return fmt.Sprintf(sc.Context, args...)
}

// renderMarkdown renders src through goldmark with the same extension
// chain gnoweb uses in production (NewGnoExtension; no image validator
// — tests should construct cases that don't depend on validator state).
func renderMarkdown(t *testing.T, src string) string {
	t.Helper()
	gnourl, err := weburl.Parse("https://gno.land/r/test")
	require.NoError(t, err)
	ctxOpts := parser.WithContext(NewGnoParserContext(GnoContext{GnoURL: gnourl, Links: FollowInternalLinks}))
	// Mirror gnoweb's production extension chain (render_config.go:43-54)
	// so goldens exercise the same parser configuration users hit. Without
	// extension.Table, the table-related fixtures would render as plain
	// paragraphs instead of <table>, masking whether the sanitizer's
	// table-preservation / cross-paragraph-isolation claims actually hold
	// under production parsing.
	ext := NewGnoExtension()
	m := goldmark.New(
		goldmark.WithExtensions(
			extension.Strikethrough,
			extension.Table,
			extension.Footnote,
			extension.TaskList,
		),
	)
	ext.Extend(m)
	node := m.Parser().Parse(text.NewReader([]byte(src)), ctxOpts)
	var buf bytes.Buffer
	require.NoError(t, m.Renderer().Render(&buf, []byte(src), node))
	return buf.String()
}

// ----- test driver -----

func TestSanitizeIntegration(t *testing.T) {
	files := []string{}
	err := filepath.Walk(sanitizeTestdataDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(info.Name()) == ".txtar" {
			files = append(files, path)
		}
		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, files, "no .txtar cases found in %s", sanitizeTestdataDir)

	for _, file := range files {
		name, _ := strings.CutPrefix(file, filepath.Clean(sanitizeTestdataDir)+"/")
		name, _ = strings.CutSuffix(name, filepath.Ext(name))

		t.Run(name, func(t *testing.T) {
			archive, err := txtar.ParseFile(file)
			require.NoError(t, err)
			sc := parseDirectives(t, archive.Comment)

			// Locate input section (required); output / html are
			// regenerated on demand.
			var inputData, wantOutput, wantHTML []byte
			haveOutput, haveHTML := false, false
			for _, f := range archive.Files {
				switch f.Name {
				case "input.md":
					inputData = f.Data
				case "output.md":
					wantOutput = f.Data
					haveOutput = true
				case "output.html":
					wantHTML = f.Data
					haveHTML = true
				default:
					t.Fatalf("unknown section %q in %s", f.Name, file)
				}
			}
			require.NotNil(t, inputData, "missing -- input.md -- section in %s", file)

			// txtar appends a trailing newline to every section; strip
			// that one newline before sanitizing (it's a fixture artifact,
			// not part of the caller input). If INPUT_ESCAPED is set,
			// also decode Go-string escapes so the case can author CR,
			// NUL, or other bytes that editors normalize.
			input := strings.TrimSuffix(string(inputData), "\n")
			if sc.Escaped {
				input = decodeInput(t, input)
			}
			gotOutput := applySanitize(t, sc, input)

			var gotHTML string
			wantHTMLSection := sc.Context != ""
			if wantHTMLSection {
				gotHTML = renderMarkdown(t, substituteContext(sc, gotOutput))
			}

			if *update {
				archive.Files = archive.Files[:0]
				archive.Files = append(archive.Files, txtar.File{Name: "input.md", Data: inputData})
				archive.Files = append(archive.Files, txtar.File{Name: "output.md", Data: addTrailingNewline([]byte(gotOutput))})
				if wantHTMLSection {
					archive.Files = append(archive.Files, txtar.File{Name: "output.html", Data: addTrailingNewline([]byte(gotHTML))})
				}
				require.NoError(t, os.WriteFile(file, txtar.Format(archive), 0o644))
				t.Logf("updated %s", file)
				return
			}

			require.True(t, haveOutput, "missing -- output.md -- section; run with -update-golden-tests")
			require.Equal(t, strings.TrimSuffix(string(wantOutput), "\n"), strings.TrimSuffix(gotOutput, "\n"), "sanitize output mismatch")

			if wantHTMLSection {
				require.True(t, haveHTML, "CONTEXT directive present but -- output.html -- section missing; run with -update-golden-tests")
				require.Equal(t, strings.TrimSuffix(string(wantHTML), "\n"), strings.TrimSuffix(gotHTML, "\n"), "rendered HTML mismatch")
			} else {
				require.False(t, haveHTML, "no CONTEXT directive but -- output.html -- section present in %s", file)
			}
		})
	}
}

func addTrailingNewline(b []byte) []byte {
	if len(b) == 0 || b[len(b)-1] != '\n' {
		return append(b, '\n')
	}
	return b
}

// TestBlockRichSetextCannotReachRealm verifies the cross-boundary
// setext-promotion defense built into sanitize.BlockRich. A realm that
// concatenates `realm_chrome + BlockRich(user_input)` must NOT allow
// the user to promote `realm_chrome`'s last line to a heading by
// starting their content with `===` or `---`. BlockRich's
// neuterLeadingSetextIfQualifying pre-pass inserts a `\` before the
// first `=`/`-` of the first non-blank line of user input if it
// matches the CM §4.3 setext-underline pattern, which defeats the
// promotion regardless of whether the realm's tail ended in `\n`.
func TestBlockRichSetextCannotReachRealm(t *testing.T) {
	sc := sanitizeCase{Func: "BlockRich"}
	cases := []struct {
		name        string
		realmChrome string
		userInput   string
		htmlMustNot string // substring that should NOT appear in rendered HTML
	}{
		{
			name:        "realm-with-trailing-newline-plus-setext-h1",
			realmChrome: "Welcome to MyRealm\n",
			userInput:   "===\ncaller text\n",
			htmlMustNot: "<h1>Welcome to MyRealm</h1>",
		},
		{
			name:        "realm-without-trailing-newline-plus-setext-h1",
			realmChrome: "Welcome to MyRealm",
			userInput:   "===\ncaller text\n",
			htmlMustNot: "<h1>Welcome to MyRealm</h1>",
		},
		{
			name:        "realm-without-trailing-newline-plus-setext-h2",
			realmChrome: "Realm Banner",
			userInput:   "---\ncaller text\n",
			htmlMustNot: "<h2>Realm Banner</h2>",
		},
		{
			name:        "realm-without-trailing-newline-plus-leading-blank-setext",
			realmChrome: "Realm Banner",
			userInput:   "\n===\ncaller text\n",
			htmlMustNot: "<h1>Realm Banner</h1>",
		},
		{
			name:        "realm-without-trailing-newline-plus-indented-setext",
			realmChrome: "Realm Banner",
			userInput:   "   ===\ncaller text\n",
			htmlMustNot: "<h1>Realm Banner</h1>",
		},
		{
			// U+2028 line separator — folded to \n by the native, so
			// the qualifying-setext pre-pass must also recognize it.
			name:        "realm-no-trailing-newline-plus-u2028-prefix-setext",
			realmChrome: "Welcome to MyRealm",
			userInput:   "\u2028===\ncaller text\n",
			htmlMustNot: "<h1>Welcome to MyRealm</h1>",
		},
		{
			// U+2029 paragraph separator — same concern.
			name:        "realm-no-trailing-newline-plus-u2029-prefix-setext",
			realmChrome: "Realm Banner",
			userInput:   "\u2029---\ncaller text\n",
			htmlMustNot: "<h2>Realm Banner</h2>",
		},
		{
			// U+0085 NEL — same concern.
			name:        "realm-no-trailing-newline-plus-nel-prefix-setext",
			realmChrome: "Welcome to MyRealm",
			userInput:   "\u0085===\ncaller text\n",
			htmlMustNot: "<h1>Welcome to MyRealm</h1>",
		},
		{
			// Deeper-setext attack: user content's first line is text,
			// but a deeper `===` could promote a merged realm+user line
			// if the realm concatenates without a separator. The
			// leading-"\n" output defense forces user content onto a
			// new line so the realm chrome cannot be merged into the
			// promoted heading.
			name:        "realm-no-trailing-newline-plus-deeper-setext-h1",
			realmChrome: "Welcome to MyRealm",
			userInput:   "# user title\n===\nbody\n",
			htmlMustNot: "<h1>Welcome to MyRealm# user title</h1>",
		},
		{
			name:        "realm-no-trailing-newline-plus-deeper-setext-h2",
			realmChrome: "Realm Banner",
			userInput:   "user section\n---\nbody\n",
			htmlMustNot: "<h2>Realm Bannerusersection</h2>",
		},
		{
			name:        "realm-no-trailing-newline-plus-deeper-setext-merging-text",
			realmChrome: "Realm Banner",
			userInput:   "user heading\n===\nbody\n",
			htmlMustNot: "<h1>Realm Banneruser heading</h1>",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sanitized := applySanitize(t, sc, c.userInput)
			composed := c.realmChrome + sanitized
			html := renderMarkdown(t, composed)
			require.NotContains(t, html, c.htmlMustNot,
				"setext attack succeeded — realm chrome %q was promoted via user input %q\ncomposed source: %q\nrendered: %q",
				c.realmChrome, c.userInput, composed, html)
			// Idempotency: BlockRich(BlockRich(userInput)) must be
			// byte-identical to BlockRich(userInput) — the cross-
			// boundary defense must compose cleanly across repeat
			// calls.
			twice := applySanitize(t, sc, sanitized)
			require.Equal(t, sanitized, twice,
				"BlockRich not idempotent for user input %q", c.userInput)
		})
	}
}

// TestExtensionAndFooterIsolation verifies that untrusted user input cannot
// open live structural extensions (<gno-columns>, <gno-form>, <gno-frame>, etc.),
// swallow trusted page footer chrome, or rebind the realm's reference links
// across Block and BlockRich callers.
func TestExtensionAndFooterIsolation(t *testing.T) {
	const (
		chromePrefix = "# Trusted Realm Title\n\n"
		chromeFooter = "\n\n---\n**Trusted Realm Footer** [home]\n\n[home]: https://realm.example/home\n"
	)

	// A live structural extension, or a link bound to the attacker's URL.
	liveExtension := regexp.MustCompile(`class="gno-(?:form|columns|column|frame|foreign|button|alert)[" ]`)
	const rebound = `href="https://evil.example`

	reproCases := []struct{ name, input string }{
		{"repro-bypass-fence-tilde", " ~~~\n~~~\n<gno-columns>"},
		{"repro-bypass-fence-backtick", " ```\nx\n```\n<gno-columns>"},
		{"repro-fence-in-list-form", "- a\n  ```\n<gno-form path=\"x\">\n```\n"},
		{"repro-fence-in-list-blank-line", "- a\n\n  ```\n<gno-form path=\"x\">\n```\n"},
		{"repro-fence-in-list-lazy-continuation", "- a\nb\n  ```\n<gno-columns>\n```\n"},
		{"repro-fence-in-list-dedented-closer", "- a\n  ```\n```\n<gno-columns>\n"},
		{"repro-fence-in-list-unclosed", "- a\n  ```\n  code\n"},
		{"repro-fence-after-thematic-break", "para\n* * *\n  ```\n  x\ntext\n"},
		{"repro-fence-after-ordered-marker", "para\n2. item\n   ```\n   x\n<gno-columns>\n"},
		{"repro-fence-in-list-lrd", "- a\n  ```\n[home]: https://evil.example\n```\n"},
		{"repro-fence-in-list-lrd-then-closer", "- a\n  ```\n[x]: https://x.example\n```\n[home]: https://evil.example\n"},
		{"repro-indented-code-lrd", "    ```\n[home]: https://evil.example\n"},
		{"repro-tab-indented-code-lrd", "\t```\n[home]: https://evil.example\n"},
		{"repro-unclosed-fence-indented", " ```\nx\n"},
		{"bracket-fence-idempotency", "[\n ~~~](<\n>)"},
		{"repro-lrd-destination-line-opens-fence", "- x\n[foo]:\n  ```\n<gno-form path=\"x\">\n"},
		{"repro-lrd-indent-joins-next", "- x\n  [foo]: /u\n```\n<gno-form path=\"x\">\n```\n"},
		{"repro-lrd-indent-before-fence", " [x]: u\n   ```\n[home]: https://evil.example\n"},
		{"repro-empty-list-item", "-\n  ```\n<gno-form path=\"x\">\n```\n"},
		{"repro-empty-list-item-lrd", "*\n  ```\n[home]: https://evil.example\n```\n"},
		{"repro-empty-item-after-fence-footer", "```\n```\n*\n  ```\n<gno-form path=\"x\">\n```\n"},
		{"repro-alert-quote-tab-fence", "\t````\n````\n> [!WARNING] x"},
		{"repro-alert-rich-list-fence", "- x\n    ```\n  ```\n  > [!WARNING] x"},
		{"repro-alert-quote-tab-closer", "```\n\t```\n> [!WARNING] x"},
		{"repro-fence-on-list-marker-line", "- ```\n  ```\n  <gno-form path=\"x\">\n"},
		{"repro-closer-at-5-spaces", "- a\n  ```\n     ```\n  <gno-form path=\"x\">\n"},
		{"repro-opener-at-4-spaces-lrd", "- a\n    ```\n  ```\n  [home]: https://evil.example\n"},
		{"repro-html-block-hides-fence", "<div>\n```\n\n<gno-form path=\"x\">\n"},
		{"repro-hard-break-before-fence", "a\\\n```\nx\n```\n[home]: https://evil.example\n"},
		{"repro-tab-indented-line-in-list-fence", "1. a\n   ```\n\t<gno-form path=\"x\">\n"},
		{"repro-lrd-text-after-destination", "[a]: u ```\n```\n- y\n  ```\n~~~\n"},
		{"repro-lrd-text-after-destination-html", "[a]: u ```\n```\n- y\n  ```\n<script>\n"},
		{"repro-html-close-tag-space-over-fence", "</ div>\n```\n\nfoo"},
		{"repro-html-block-type-1-form-feed", "<script\f\nhello"},
		{"idempotent-empty-item-in-link", "[](\n1)"},
		{"idempotent-leading-setext-after-lrd", " []: x\n   ="},
	}

	for _, helper := range []string{"Block", "BlockRich", "Blockquote", "BlockquoteRich"} {
		sc := sanitizeCase{Func: helper}
		t.Run(helper, func(t *testing.T) {
			for _, rc := range reproCases {
				t.Run(rc.name, func(t *testing.T) {
					sanitized := applySanitize(t, sc, rc.input)
					composed := chromePrefix + sanitized + chromeFooter
					html := renderMarkdown(t, composed)

					require.False(t, liveExtension.MatchString(html),
						"helper %s let input %q open an extension\nrendered: %s",
						helper, rc.input, html)
					require.NotContains(t, html, rebound,
						"helper %s let input %q rebind [home]\nrendered: %s",
						helper, rc.input, html)
					require.Contains(t, html, `href="https://realm.example/home"`,
						"helper %s let input %q drop or corrupt realm link [home]\nrendered: %s",
						helper, rc.input, html)

					// The footer must render as markdown: inside a code
					// block its text would still be there, unbolded.
					require.Contains(t, html, "<strong>Trusted Realm Footer</strong>",
						"trusted footer was swallowed or corrupted by helper %s with input %q\nrendered: %s",
						helper, rc.input, html)

					// The quote helpers nest a second `> ` on a second
					// pass by design; the others must be idempotent.
					if helper == "Block" || helper == "BlockRich" {
						twice := applySanitize(t, sc, sanitized)
						require.Equal(t, sanitized, twice,
							"helper %s was not idempotent on input %q", helper, rc.input)
					}
				})
			}
		})
	}

	// Verify that <gno-icon> inside list items and blockquotes renders as a live SVG in BlockRich
	t.Run("BlockRich_IconInContainers", func(t *testing.T) {
		sc := sanitizeCase{Func: "BlockRich"}
		containerCases := []struct {
			name  string
			input string
		}{
			{"list-bullet", "- <gno-icon name=\"star\" /> Star item\n"},
			{"list-numbered", "1. <gno-icon name=\"check\" /> Checked item\n"},
			{"blockquote", "> <gno-icon name=\"info\" /> Note quote\n"},
		}
		for _, tc := range containerCases {
			t.Run(tc.name, func(t *testing.T) {
				sanitized := applySanitize(t, sc, tc.input)
				html := renderMarkdown(t, sanitized)
				require.Contains(t, html, "<svg class=\"gno-icon\"",
					"gno-icon in container failed to render live SVG: %s", html)

				// Idempotency
				twice := applySanitize(t, sc, sanitized)
				require.Equal(t, sanitized, twice, "not idempotent on %q", tc.input)
			})
		}
	})
}

// SANITIZE.md's threat-surface table carries a per-helper fixture count. Hand
// maintenance let it drift badly — it claimed 103 fixtures against 186 on
// disk, documented a `Link` helper that no longer exists, and had no row at
// all for BlockRich (30 fixtures) or BlockquoteRich (1). A stale count reads
// as coverage that is not there, which is the opposite of what the corpus is
// for, so the numbers are asserted rather than trusted. The Threats column
// stays prose and is deliberately unchecked.
func TestSanitizeCoverageTableMatchesCorpus(t *testing.T) {
	corpus := map[string]int{}
	err := filepath.Walk(sanitizeTestdataDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(info.Name()) != ".txtar" {
			return nil
		}
		archive, err := txtar.ParseFile(path)
		if err != nil {
			return err
		}
		sc := parseDirectives(t, archive.Comment)
		require.NotEmpty(t, sc.Func, "%s: missing MARKDOWNFUNC directive", path)
		corpus[sc.Func]++
		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, corpus)

	doc, err := os.ReadFile("SANITIZE.md")
	require.NoError(t, err)

	// Rows look like: | `Helper` | 12 | threats… |
	rowRe := regexp.MustCompile("(?m)^\\| `([A-Za-z]+)` \\| ([0-9]+) \\|")
	documented := map[string]int{}
	for _, m := range rowRe.FindAllStringSubmatch(string(doc), -1) {
		n, err := strconv.Atoi(m[2])
		require.NoError(t, err)
		documented[m[1]] = n
	}
	require.NotEmpty(t, documented, "no helper rows parsed out of SANITIZE.md")

	for fn, want := range corpus {
		got, ok := documented[fn]
		require.True(t, ok,
			"SANITIZE.md has no row for %s, which has %d fixtures", fn, want)
		require.Equal(t, want, got,
			"SANITIZE.md row for %s says %d cases, corpus has %d", fn, got, want)
	}
	for fn := range documented {
		require.Contains(t, corpus, fn,
			"SANITIZE.md documents %s, which has no fixtures (helper removed?)", fn)
	}

	total := 0
	for _, n := range corpus {
		total += n
	}
	require.Contains(t, string(doc), fmt.Sprintf("%d fixtures total", total),
		"SANITIZE.md total is stale; corpus has %d fixtures", total)
}
