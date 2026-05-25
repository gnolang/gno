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
//	-- input.md --     user-supplied content (the attacker input)
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
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
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

// Base stores are initialized once and shared across cases via per-case
// CacheWrap + BeginTransaction layering (same pattern filetests use to
// keep cases isolated from each other's package writes).
var (
	baseStoreOnce sync.Once
	baseCommit    storetypes.CommitStore
	baseGno       gno.Store
)

func sanitizeBaseStores(t *testing.T) (storetypes.CommitStore, gno.Store) {
	t.Helper()
	baseStoreOnce.Do(func() {
		baseCommit, baseGno = test.TestStore(gnoenv.RootDir(), io.Discard, nil)
	})
	return baseCommit, baseGno
}

// callSanitizeGno spawns a fresh Machine for one case, loads the driver
// package into it (which itself imports sanitize/v0), and Evals
// `Run(fn, input, arg)`. The Machine is released before return.
//
// Per-case isolation: wrap the base commit store, build a fresh
// transactional gno store on top of that wrap, and never commit. This
// matches the filetest pattern and prevents one case's package
// registrations from leaking into the next.
func callSanitizeGno(t *testing.T, fn, input, arg, arg2 string) string {
	t.Helper()
	commit, gnoBase := sanitizeBaseStores(t)
	tcw := commit.CacheWrap()
	tx := gnoBase.BeginTransaction(tcw, tcw, nil, nil)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		Store:         tx,
		Context:       test.Context("", driverPkgPath, nil),
		Output:        io.Discard,
		MaxAllocBytes: math.MaxInt64,
	})
	defer m.Release()

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
	args := make([]interface{}, count)
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
	ctxOpts := parser.WithContext(NewGnoParserContext(GnoContext{GnoURL: gnourl}))
	ext := NewGnoExtension()
	m := goldmark.New()
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
			// not part of the attacker input). If INPUT_ESCAPED is set,
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
