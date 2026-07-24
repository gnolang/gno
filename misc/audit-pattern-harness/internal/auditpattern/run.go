package auditpattern

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/format"
	"go/scanner"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var exportedPointerVarRE = regexp.MustCompile(`^var\s+[A-Z]\w*\s+\*`)
var exportedPointerFuncRE = regexp.MustCompile(`^func\s+([A-Z]\w*)\([^)]*\)\s+\*`)
var freshConstructorReturnRE = regexp.MustCompile(`return\s+&[A-Z]\w*\s*\{`)
var mapVarRE = regexp.MustCompile(`^(?:var\s+)?([A-Za-z_]\w*)\s*(?:=\s*)?map\[`)

type Options struct {
	GNOBin string
}

type Report struct {
	ID       string          `json:"id"`
	Title    string          `json:"title"`
	Rule     string          `json:"rule"`
	OK       bool            `json:"ok"`
	Fixtures []FixtureResult `json:"fixtures"`
}

type FixtureResult struct {
	Name                 string   `json:"name"`
	Path                 string   `json:"path"`
	PathOK               bool     `json:"path_ok"`
	GNOTestOK            bool     `json:"gno_test_ok"`
	GNOTestWant          string   `json:"gno_test_want"`
	GNOTestOutput        string   `json:"gno_test_output"`
	PatternHits          []Hit    `json:"pattern_hits"`
	WantPatternHits      int      `json:"want_pattern_hits"`
	PatternExpectationOK bool     `json:"pattern_expectation_ok"`
	Errors               []string `json:"errors,omitempty"`
}

type Hit struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

func Run(ctx context.Context, rec Record, opts Options) Report {
	report := Report{
		ID:    rec.ID,
		Title: rec.Title,
		Rule:  rec.Rule,
		OK:    true,
	}

	for _, fixture := range rec.Fixtures {
		result := runFixture(ctx, rec.Rule, fixture, opts)
		if len(result.Errors) > 0 || !result.PathOK || !result.GNOTestOK || !result.PatternExpectationOK {
			report.OK = false
		}
		report.Fixtures = append(report.Fixtures, result)
	}

	return report
}

func runFixture(ctx context.Context, rule string, fixture Fixture, opts Options) FixtureResult {
	result := FixtureResult{
		Name:            fixture.Name,
		Path:            fixture.Path,
		GNOTestWant:     fixture.WantGNOTest,
		WantPatternHits: fixture.WantPatternHits,
	}

	info, err := os.Stat(fixture.Path)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result
	}
	if !info.IsDir() {
		result.Errors = append(result.Errors, "fixture path is not a directory")
		return result
	}
	result.PathOK = true

	testPass, output := runGNOTest(ctx, opts.GNOBin, fixture.Path)
	result.GNOTestOutput = output
	result.GNOTestOK = (fixture.WantGNOTest == "pass" && testPass) || (fixture.WantGNOTest == "fail" && !testPass)

	hits, err := RunRule(rule, fixture.Path)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
	}
	result.PatternHits = hits
	result.PatternExpectationOK = len(hits) == fixture.WantPatternHits

	return result
}

func runGNOTest(ctx context.Context, gnoBin, dir string) (bool, string) {
	if gnoBin == "" {
		gnoBin = "gno"
	}
	cmd := exec.CommandContext(ctx, gnoBin, "test", ".")
	cmd.Dir = dir
	cmd.Env = os.Environ()
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	output := strings.TrimSpace(out.String())
	if ctx.Err() != nil {
		return false, "[timeout] " + output
	}
	return err == nil, output
}

func RunRule(rule, dir string) ([]Hit, error) {
	switch rule {
	case "current_guard":
		return currentGuardHits(dir)
	case "render_markdown_escape":
		return renderMarkdownEscapeHits(dir)
	case "payment_user_call":
		return paymentUserCallHits(dir)
	case "origin_caller_auth":
		return originCallerAuthHits(dir)
	case "callback_param":
		return callbackParamHits(dir)
	case "interface_realm_param":
		return interfaceRealmParamHits(dir)
	case "exported_pointer_leak":
		return exportedPointerLeakHits(dir)
	case "render_map_iteration":
		return renderMapIterationHits(dir)
	default:
		return nil, fmt.Errorf("unknown rule %q", rule)
	}
}

func currentGuardHits(dir string) ([]Hit, error) {
	files, err := gnoFiles(dir)
	if err != nil {
		return nil, err
	}

	var hits []Hit
	for _, file := range files {
		data, err := readGnoSource(file)
		if err != nil {
			return nil, err
		}
		inFunc := false
		braceDepth := 0
		seenIsCurrent := false
		orig := strings.Split(string(data), "\n")
		for i, line := range codeLines(data) {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "func ") {
				inFunc = true
				braceDepth = 0
				seenIsCurrent = false
			}
			if inFunc {
				braceDepth += strings.Count(line, "{")
				braceDepth -= strings.Count(line, "}")
			}
			if strings.Contains(line, ".IsCurrent()") {
				seenIsCurrent = true
			}
			if strings.Contains(line, ".Previous()") && !seenIsCurrent {
				hits = append(hits, newHit(dir, file, i+1, orig[i]))
			}
			if inFunc && braceDepth <= 0 {
				inFunc = false
				seenIsCurrent = false
			}
		}
	}
	return hits, nil
}

func renderMarkdownEscapeHits(dir string) ([]Hit, error) {
	files, err := gnoFiles(dir)
	if err != nil {
		return nil, err
	}

	var hits []Hit
	for _, file := range files {
		data, err := readGnoSource(file)
		if err != nil {
			return nil, err
		}
		inRender := false
		braceDepth := 0
		orig := strings.Split(string(data), "\n")
		for i, line := range codeLines(data) {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "func Render(") {
				inRender = true
				braceDepth = 0
			}
			if inRender {
				braceDepth += strings.Count(line, "{")
				braceDepth -= strings.Count(line, "}")
				lower := strings.ToLower(line)
				if strings.Contains(line, "return") && strings.Contains(line, "path") && !strings.Contains(lower, "escape") {
					hits = append(hits, newHit(dir, file, i+1, orig[i]))
				}
			}
			if inRender && braceDepth <= 0 {
				inRender = false
			}
		}
	}
	return hits, nil
}

func paymentUserCallHits(dir string) ([]Hit, error) {
	files, err := gnoFiles(dir)
	if err != nil {
		return nil, err
	}

	var hits []Hit
	for _, file := range files {
		data, err := readGnoSource(file)
		if err != nil {
			return nil, err
		}
		inFunc := false
		braceDepth := 0
		seenUserCall := false
		orig := strings.Split(string(data), "\n")
		for i, line := range codeLines(data) {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "func ") {
				inFunc = true
				braceDepth = 0
				seenUserCall = false
			}
			if inFunc {
				braceDepth += strings.Count(line, "{")
				braceDepth -= strings.Count(line, "}")
			}
			if strings.Contains(line, ".IsUserCall()") {
				seenUserCall = true
			}
			if strings.Contains(line, "OriginSend()") && !seenUserCall {
				hits = append(hits, newHit(dir, file, i+1, orig[i]))
			}
			if inFunc && braceDepth <= 0 {
				inFunc = false
				seenUserCall = false
			}
		}
	}
	return hits, nil
}

func originCallerAuthHits(dir string) ([]Hit, error) {
	return lineContainsHits(dir, func(line string) bool {
		trimmed := strings.TrimSpace(line)
		return !strings.HasPrefix(trimmed, "//") &&
			strings.Contains(line, "OriginCaller()") &&
			!strings.Contains(line, "SetOriginCaller") &&
			(strings.Contains(line, "==") || strings.Contains(line, "!="))
	})
}

func callbackParamHits(dir string) ([]Hit, error) {
	// Use the original (non-trimmed) line so that function literals assigned
	// inside a body (which are always indented) are not matched as top-level
	// function declarations that accept callback parameters.
	return lineContainsHits(dir, func(line string) bool {
		return strings.HasPrefix(line, "func ") && strings.Contains(line, " func(")
	})
}

func interfaceRealmParamHits(dir string) ([]Hit, error) {
	files, err := gnoFiles(dir)
	if err != nil {
		return nil, err
	}

	var hits []Hit
	for _, file := range files {
		data, err := readGnoSource(file)
		if err != nil {
			return nil, err
		}
		inInterface := false
		braceDepth := 0
		orig := strings.Split(string(data), "\n")
		for i, line := range codeLines(data) {
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, "interface {") {
				inInterface = true
				braceDepth = 0
			}
			if inInterface {
				braceDepth += strings.Count(line, "{")
				braceDepth -= strings.Count(line, "}")
				if strings.Contains(line, "realm") {
					hits = append(hits, newHit(dir, file, i+1, orig[i]))
				}
			}
			if inInterface && braceDepth <= 0 {
				inInterface = false
			}
		}
	}
	return hits, nil
}

func exportedPointerLeakHits(dir string) ([]Hit, error) {
	files, err := gnoFiles(dir)
	if err != nil {
		return nil, err
	}

	var hits []Hit
	for _, file := range files {
		data, err := readGnoSource(file)
		if err != nil {
			return nil, err
		}
		orig := strings.Split(string(data), "\n")
		lines := codeLines(data)
		for i := range lines {
			line := lines[i]
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue
			}
			if exportedPointerVarRE.MatchString(trimmed) {
				hits = append(hits, newHit(dir, file, i+1, orig[i]))
				continue
			}
			match := exportedPointerFuncRE.FindStringSubmatch(trimmed)
			if match == nil {
				continue
			}
			if strings.HasPrefix(match[1], "New") && returnsFreshPointer(lines[i:]) {
				continue
			}
			hits = append(hits, newHit(dir, file, i+1, orig[i]))
		}
	}
	return hits, nil
}

func returnsFreshPointer(lines []string) bool {
	braceDepth := 0
	for i, line := range lines {
		braceDepth += strings.Count(line, "{")
		braceDepth -= strings.Count(line, "}")
		if freshConstructorReturnRE.MatchString(line) {
			return true
		}
		if i > 0 && braceDepth <= 0 {
			return false
		}
	}
	return false
}

func renderMapIterationHits(dir string) ([]Hit, error) {
	files, err := gnoFiles(dir)
	if err != nil {
		return nil, err
	}

	var hits []Hit
	for _, file := range files {
		data, err := readGnoSource(file)
		if err != nil {
			return nil, err
		}

		mapVars := make(map[string]struct{})
		orig := strings.Split(string(data), "\n")
		lines := codeLines(data)
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue
			}
			if match := mapVarRE.FindStringSubmatch(trimmed); match != nil {
				mapVars[match[1]] = struct{}{}
			}
		}

		inRender := false
		braceDepth := 0
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "func Render(") {
				inRender = true
				braceDepth = 0
			}
			if inRender {
				braceDepth += strings.Count(line, "{")
				braceDepth -= strings.Count(line, "}")
				if strings.Contains(line, "range ") {
					// Normalize to remove extra spaces: range     var -> range var
					normalized := strings.Join(strings.Fields(line), " ")
					for name := range mapVars {
						if strings.Contains(normalized, "range "+name) {
							hits = append(hits, newHit(dir, file, i+1, orig[i]))
						}
					}
				}
			}
			if inRender && braceDepth <= 0 {
				inRender = false
			}
		}
	}
	return hits, nil
}

func lineContainsHits(dir string, match func(string) bool) ([]Hit, error) {
	files, err := gnoFiles(dir)
	if err != nil {
		return nil, err
	}

	var hits []Hit
	for _, file := range files {
		data, err := readGnoSource(file)
		if err != nil {
			return nil, err
		}
		orig := strings.Split(string(data), "\n")
		for i, line := range codeLines(data) {
			if match(line) {
				hits = append(hits, newHit(dir, file, i+1, orig[i]))
			}
		}
	}
	return hits, nil
}

func newHit(dir, file string, line int, text string) Hit {
	rel, err := filepath.Rel(dir, file)
	if err != nil {
		rel = file
	}
	return Hit{
		File: rel,
		Line: line,
		Text: strings.TrimSpace(text),
	}
}

// readGnoSource reads a .gno file and returns its gofmt-normalized contents so
// the line-based pattern matchers are not defeated by irregular spacing. For
// example "func GetVault()*Vault{" is normalized to "func GetVault() *Vault {"
// before scanning, which the matchers expect. .gno files use Go syntax, so
// go/format applies. If the source cannot be parsed (e.g. intentionally broken
// fixture), the raw bytes are returned unchanged.
func readGnoSource(file string) ([]byte, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	formatted, err := format.Source(data)
	if err != nil {
		return data, nil
	}
	return formatted, nil
}

// codeLines splits gno source into lines with the contents of string/char
// literals and the bodies of comments blanked out (replaced with spaces),
// leaving delimiters and line structure intact. The line-based matchers run
// detection against this "code view" so that braces, keywords, or call
// expressions appearing inside a string or comment cannot fool them — e.g. a
// "}" in a string literal must not flip brace-depth tracking and turn a
// correctly guarded function into a false positive. Hits are still reported
// against the original source text. The returned slice has the same length as
// strings.Split(data, "\n").
func codeLines(data []byte) []string {
	blanked := append([]byte(nil), data...)
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(data))
	var s scanner.Scanner
	s.Init(file, data, nil, scanner.ScanComments)
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok != token.COMMENT && tok != token.STRING && tok != token.CHAR {
			continue
		}
		start := fset.Position(pos).Offset
		lo, hi := start, start+len(lit)
		if tok != token.COMMENT {
			lo, hi = start+1, hi-1 // preserve the surrounding quotes/backticks
		}
		for i := lo; i < hi && i < len(blanked); i++ {
			if blanked[i] != '\n' {
				blanked[i] = ' '
			}
		}
	}
	return strings.Split(string(blanked), "\n")
}

func gnoFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".gno" {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func (report Report) Markdown() string {
	var b strings.Builder
	status := "PASS"
	if !report.OK {
		status = "FAIL"
	}
	fmt.Fprintf(&b, "# Audit Pattern Harness: %s\n\n", report.Title)
	fmt.Fprintf(&b, "- id: `%s`\n", report.ID)
	fmt.Fprintf(&b, "- rule: `%s`\n", report.Rule)
	fmt.Fprintf(&b, "- status: `%s`\n\n", status)
	for _, fixture := range report.Fixtures {
		fixtureStatus := "PASS"
		if len(fixture.Errors) > 0 || !fixture.PathOK || !fixture.GNOTestOK || !fixture.PatternExpectationOK {
			fixtureStatus = "FAIL"
		}
		fmt.Fprintf(&b, "## %s: %s\n\n", fixture.Name, fixtureStatus)
		fmt.Fprintf(&b, "- path: `%s`\n", fixture.Path)
		fmt.Fprintf(&b, "- gno test: want `%s`, ok `%t`\n", fixture.GNOTestWant, fixture.GNOTestOK)
		fmt.Fprintf(&b, "- pattern hits: got `%d`, want `%d`\n", len(fixture.PatternHits), fixture.WantPatternHits)
		for _, hit := range fixture.PatternHits {
			fmt.Fprintf(&b, "  - `%s:%d` `%s`\n", hit.File, hit.Line, hit.Text)
		}
		for _, msg := range fixture.Errors {
			fmt.Fprintf(&b, "- error: `%s`\n", msg)
		}
		if fixture.GNOTestOutput != "" {
			fmt.Fprintf(&b, "\n```text\n%s\n```\n", fixture.GNOTestOutput)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func ReportsJSON(reports []Report) ([]byte, error) {
	return json.MarshalIndent(reports, "", "  ")
}
