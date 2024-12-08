package coverage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "github.com/gnolang/gno/tm2/pkg/commands"
)

type ansiColor string

const (
	Reset  ansiColor = "\033[0m"
	Green  ansiColor = "\033[32m"
	Yellow ansiColor = "\033[33m"
	Red    ansiColor = "\033[31m"
	White  ansiColor = "\033[37m"
	Orange ansiColor = "\033[38;5;208m"
	Bold   ansiColor = "\033[1m"
)

// color scheme for coverage report
// ColorScheme defines ANSI color codes for terminal output
type ColorScheme struct {
	Reset     ansiColor
	Success   ansiColor
	Warning   ansiColor
	Error     ansiColor
	Info      ansiColor
	Highlight ansiColor
	Bold      ansiColor
}

var defaultColors = ColorScheme{
	Reset:     Reset,
	Success:   Green,
	Warning:   Yellow,
	Error:     Red,
	Info:      White,
	Highlight: Orange,
	Bold:      Bold,
}

type ReportFormat string

const (
	Text ReportFormat = "text"
	JSON ReportFormat = "json"
	HTML ReportFormat = "html"
)

type Reporter interface {
	// Write writes a coverage report to the given writer.
	Write(w io.Writer) error

	// WriteFileDetail writes detailed coverage info for specific file.
	WriteFileDetail(w io.Writer, pattern string, showHits bool) error
}

type ReportOpts struct {
	format   ReportFormat
	showHits bool
	fileName string
	pattern  string
}

type baseReporter struct {
	coverage *Coverage
	colors   ColorScheme
}

func (base *baseReporter) sortFiles() []string {
	files := make([]string, 0, len(base.coverage.files))
	for file := range base.coverage.files {
		files = append(files, file)
	}
	sort.Strings(files)
	return files
}

func (r *baseReporter) calculateStats(cov FileCoverage) (int, int, float64) {
	executableLines := 0
	for _, executable := range cov.executableLines {
		if executable {
			executableLines++
		}
	}

	hitLines := len(cov.hitLines)
	percentage := float64(0)
	if executableLines > 0 {
		percentage = float64(hitLines) / float64(executableLines) * 100
	}

	return hitLines, executableLines, percentage
}

type ConsoleReporter struct {
	baseReporter
	finder PathFinder
}

func NewConsoleReporter(c *Coverage, finder PathFinder) *ConsoleReporter {
	return &ConsoleReporter{
		baseReporter: baseReporter{
			coverage: c,
			colors:   defaultColors,
		},
		finder: finder,
	}
}

func (r *ConsoleReporter) Write(w io.Writer) error {
	files := r.sortFiles()

	for _, file := range files {
		cov, exists := r.coverage.files.get(file)
		if !exists {
			continue
		}

		hits, executable, pct := r.calculateStats(cov)
		if executable == 0 {
			continue
		}

		color := r.colorize(r.colors, pct)
		_, err := fmt.Fprintf(w,
			"%s%.1f%% [%4d/%d] %s%s\n",
			color, floor1(pct), hits, executable, file, r.colors.Reset,
		)
		if err != nil {
			return fmt.Errorf("writing coverage for %s: %w", file, err)
		}
	}

	return nil
}

func (r *ConsoleReporter) WriteFileDetail(w io.Writer, pattern string, showHits bool) error {
	files := findMatchingFiles(r.coverage.files, pattern)
	if len(files) == 0 {
		return fmt.Errorf("no files found matching pattern %s", pattern)
	}

	for _, path := range files {
		absPath, err := r.finder.Find(path)
		if err != nil {
			return fmt.Errorf("finding file path: %w", err)
		}

		relPath := path

		if err := r.writeFileCoverage(w, absPath, relPath, showHits); err != nil {
			return err
		}

		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	return nil
}

func (r *ConsoleReporter) writeFileCoverage(w io.Writer, absPath, relPath string, showHits bool) error {
	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	// file name
	if _, err := fmt.Fprintf(w, "%s%s%s:\n", r.colors.Bold, relPath, r.colors.Reset); err != nil {
		return err
	}

	cov, exists := r.coverage.files.get(relPath)
	if !exists {
		return fmt.Errorf("no coverage data for file %s", relPath)
	}

	// print file content (line by line)
	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		line := scanner.Text()
		hits, covered := cov.hitLines[lineNum]
		executable := cov.executableLines[lineNum]

		lineInfo := r.formatLineInfo(lineNum, line, hits, covered, executable, showHits)
		if _, err := fmt.Fprintln(w, lineInfo); err != nil {
			return err
		}
		lineNum++
	}

	return scanner.Err()
}

func (r *ConsoleReporter) formatLineInfo(
	lineNum int,
	line string,
	hits int,
	covered, executable, showHits bool,
) string {
	lineNumStr := fmt.Sprintf("%4d", lineNum)
	color := r.getLineColor(covered, executable)
	hitInfo := r.formatHitInfo(hits, covered, showHits)

	return fmt.Sprintf("%s%s%s %s%s%s%s",
		color, lineNumStr, r.colors.Reset,
		hitInfo, color, line, r.colors.Reset)
}

func (r *ConsoleReporter) formatHitInfo(hits int, covered, showHits bool) string {
	if !showHits {
		return ""
	}
	if covered {
		return fmt.Sprintf("%s%-4d%s ", r.colors.Highlight, hits, r.colors.Reset)
	}
	return strings.Repeat(" ", 5)
}

func (r *ConsoleReporter) getLineColor(covered, executable bool) ansiColor {
	switch {
	case covered:
		return r.colors.Success
	case executable:
		return r.colors.Warning
	default:
		return r.colors.Info
	}
}

func (r *ConsoleReporter) colorize(scheme ColorScheme, pct float64) ansiColor {
	switch {
	case pct >= 80:
		return scheme.Success
	case pct >= 50:
		return scheme.Warning
	default:
		return scheme.Error
	}
}

type PathFinder interface {
	Find(path string) (string, error)
}

type defaultPathFinder struct {
	rootDir string
	cache   map[string]string
}

func NewDefaultPathFinder(rootDir string) PathFinder {
	return &defaultPathFinder{
		rootDir: rootDir,
		cache:   make(map[string]string),
	}
}

func (f *defaultPathFinder) Find(path string) (string, error) {
	if cached, ok := f.cache[path]; ok {
		return cached, nil
	}

	// try direct path first
	direct := filepath.Join(f.rootDir, path)
	if _, err := os.Stat(direct); err == nil {
		f.cache[path] = direct
		return direct, nil
	}

	var found string
	err := filepath.WalkDir(f.rootDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Base(p) == filepath.Base(path) {
			found = p
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("finding path %s: %w", path, err)
	}

	if found == "" {
		return "", fmt.Errorf("file %s not found", path)
	}

	f.cache[path] = found
	return found, nil
}

type JSONReporter struct {
	baseReporter
	fileName string
}

type jsonCoverage struct {
	Files map[string]jsonFileCoverage `json:"files"`
}

type jsonFileCoverage struct {
	TotalLines int            `json:"total_lines"`
	HitLines   map[string]int `json:"hit_lines"`
}

func NewJSONReporter(cov *Coverage, fname string) *JSONReporter {
	return &JSONReporter{
		baseReporter: baseReporter{coverage: cov},
		fileName:     fname,
	}
}

func (r *JSONReporter) Write(w io.Writer) error {
	data := jsonCoverage{
		Files: make(map[string]jsonFileCoverage),
	}

	for file, coverage := range r.coverage.files {
		hits := make(map[string]int)
		for line, count := range coverage.hitLines {
			hits[strconv.Itoa(line)] = count
		}

		data.Files[file] = jsonFileCoverage{
			TotalLines: coverage.totalLines,
			HitLines:   hits,
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (r *JSONReporter) WriteFileDetail(w io.Writer, pattern string, showHits bool) error {
	return fmt.Errorf("file detail view not supported for JSON format")
}

func NewReporter(cov *Coverage, opts ReportOpts) Reporter {
	switch opts.format {
	case JSON:
		return NewJSONReporter(cov, opts.fileName)
	case HTML:
		// TODO: implement HTML reporter
		return nil
	default:
		return NewConsoleReporter(cov, NewDefaultPathFinder(cov.rootDir))
	}
}

func floor1(v float64) float64 {
	return math.Floor(v*10) / 10
}
