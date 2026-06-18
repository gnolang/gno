package auditloop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

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
	GNOTestOutput        string   `json:"gno_test_output,omitempty"`
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
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return err == nil, strings.TrimSpace(out.String())
}

func RunRule(rule, dir string) ([]Hit, error) {
	switch rule {
	case "current_guard":
		return currentGuardHits(dir)
	case "render_markdown_escape":
		return renderMarkdownEscapeHits(dir)
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
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		inFunc := false
		braceDepth := 0
		seenIsCurrent := false
		for i, line := range strings.Split(string(data), "\n") {
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
				hits = append(hits, newHit(dir, file, i+1, line))
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
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		inRender := false
		braceDepth := 0
		for i, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "func Render(") {
				inRender = true
				braceDepth = 0
			}
			if inRender {
				braceDepth += strings.Count(line, "{")
				braceDepth -= strings.Count(line, "}")
				if strings.Contains(line, "return") && strings.Contains(line, "path") && !strings.Contains(line, "escape") {
					hits = append(hits, newHit(dir, file, i+1, line))
				}
			}
			if inRender && braceDepth <= 0 {
				inRender = false
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
	fmt.Fprintf(&b, "# Audit Loop: %s\n\n", report.Title)
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
