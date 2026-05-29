package test

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Errorcheck-style support for .go filetests under tests/files/gocorpus/testdata/.
//
// When a .go filetest carries inline `// ERROR "regex"` markers (Go's
// standard test corpus convention) and no explicit `// Error:` directive,
// the harness:
//   1. Parses the markers (regex patterns per line).
//   2. Applies a PKGPATH + synthetic-main rescue so files declaring a
//      non-main package (typical for declaration-level errorcheck tests)
//      reach Gno's preprocess+typecheck instead of bailing on the
//      realm-naming requirement.
//   3. Runs the file through Gno and captures its preprocess /
//      typecheck / runtime error output.
//   4. Verifies at least one marker's regex matches that output.
//
// Pass criterion is intentionally loose: Gno's preprocessor stops at
// the first error, so strict per-line marker matching would fail most
// corpus errorcheck files. We want the verdict-level signal to be
// "Gno catches the kind of error gc does", not "Gno enumerates every
// individual error" — which is too strict given the early-bail
// behaviour. See [VerifyErrorcheckMarkers] for the three outcomes.
//
// To bless an intentional divergence (Gno doesn't reject a file Go's
// errorcheck flags, or Gno's wording diverges entirely), add an
// explicit `// Error:` directive with Gno's actual output — the
// marker verification is then bypassed and the directive serves as
// documentation of the accepted divergence.
//
// The exported entry points ([HasInlineErrorMarkers],
// [ParseInlineErrors], [MarkerMatches], [VerifyErrorcheckMarkers],
// [IsRunnable], [PrependPkgPathIfNeeded]) are intended for use by
// external test drivers that walk a Go test corpus and dispatch
// files through this harness. The corresponding internal dispatch
// path in runFiletest uses these same helpers.

// InlineError is one `// ERROR "regex"` marker attached to a source
// line. Patterns are kept as Go regex strings; alternatives inside a
// pattern are separated by '|' per Go's errorcheck convention.
type InlineError struct {
	// Line is the 1-based source line the marker was attached to.
	Line int
	// Patterns lists the quoted regex strings on the marker, in
	// declaration order. Empty for a bare `// ERROR` with no patterns.
	Patterns []string
}

// HasInlineErrorMarkers reports whether source contains any
// `// ERROR ` or `// GC_ERROR ` marker. Cheap pre-check intended for
// dispatch logic that decides whether to enter errorcheck mode
// without paying for full marker parsing first.
func HasInlineErrorMarkers(source []byte) bool {
	for _, tag := range []string{"// ERROR ", "// GC_ERROR ", "// ERROR\t", "// GC_ERROR\t"} {
		if bytes.Contains(source, []byte(tag)) {
			return true
		}
	}
	return false
}

// ParseInlineErrors scans source for inline `// ERROR "..."` markers
// and returns one [InlineError] per source line that carries one.
//
// Format: any number of double-quoted (or backtick-quoted) strings
// after `// ERROR`, e.g. `// ERROR "oct|char"` or `// ERROR "a" "b"`.
// `// GC_ERROR` is accepted as equivalent to `// ERROR` since several
// corpus files mix the two; gccgo-only markers (`// GCCGO_ERROR`)
// are intentionally ignored — this harness mirrors gc semantics.
// LINE/LINE+N substitutions are NOT performed; the literal text
// stays in the regex (matching is best-effort).
//
// Lines that begin with `//` (i.e. pure doc-comment lines that
// happen to mention `// ERROR "..."` in their prose) are skipped —
// real markers always trail actual code on the same line.
func ParseInlineErrors(source []byte) []InlineError {
	var out []InlineError
	sc := bufio.NewScanner(bytes.NewReader(source))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		text := sc.Text()
		idx := indexErrorMarker(text)
		if idx < 0 {
			continue
		}
		// Skip doc-comment lines: an `// ERROR "..."` inside narrative
		// prose is not a real marker. Detect by checking that the
		// content before the marker is non-empty after stripping
		// leading whitespace — a real marker is preceded by code.
		if strings.HasPrefix(strings.TrimLeft(text, " \t"), "//") {
			continue
		}
		patterns := extractQuotedStrings(text[idx:])
		out = append(out, InlineError{Line: lineNo, Patterns: patterns})
	}
	return out
}

// indexErrorMarker returns the byte index in line where an `// ERROR`
// or `// GC_ERROR` marker's content begins, or -1.
func indexErrorMarker(line string) int {
	for _, tag := range []string{"// ERROR ", "// GC_ERROR ", "// ERROR\t", "// GC_ERROR\t"} {
		if i := strings.Index(line, tag); i >= 0 {
			return i + len(tag)
		}
	}
	return -1
}

// extractQuotedStrings pulls every Go-syntax quoted string (double
// or backtick) out of s, decoding escapes via [strconv.Unquote].
func extractQuotedStrings(s string) []string {
	var out []string
	for i := 0; i < len(s); {
		c := s[i]
		if c != '"' && c != '`' {
			i++
			continue
		}
		end := findClosingQuote(s, i)
		if end < 0 {
			break
		}
		if v, err := strconv.Unquote(s[i : end+1]); err == nil {
			out = append(out, v)
		}
		i = end + 1
	}
	return out
}

func findClosingQuote(s string, start int) int {
	open := s[start]
	if open == '`' {
		i := strings.IndexByte(s[start+1:], '`')
		if i < 0 {
			return -1
		}
		return start + 1 + i
	}
	for i := start + 1; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			continue
		}
		if s[i] == '"' {
			return i
		}
	}
	return -1
}

// MarkerMatches reports whether any of m.Patterns (or any '|'
// alternative inside one) matches actualErr as a Go regex. A bare
// marker with no patterns is satisfied by any non-empty error.
func MarkerMatches(m InlineError, actualErr string) bool {
	if len(m.Patterns) == 0 {
		return true
	}
	for _, p := range m.Patterns {
		for alt := range strings.SplitSeq(p, "|") {
			re, err := regexp.Compile(alt)
			if err != nil {
				continue
			}
			if re.MatchString(actualErr) {
				return true
			}
		}
	}
	return false
}

// VerifyErrorcheckMarkers compares markers against the concatenation
// of Gno's runtime/preprocess error output and any go/types output.
// Returns nil on PASS; an error describing the divergence otherwise.
//
// Outcomes:
//   - PASS: at least one marker's regex matched. Partial matches
//     (Gno emits one error matching one of N markers) count as PASS
//     since Gno's early-bail makes strict per-marker matching too
//     strict for any meaningful number of corpus files.
//   - FAIL "Gno accepted file": actualErr is empty; Gno was more
//     lenient than gc. Real conformance signal — likely a bug.
//   - FAIL "no marker matched": actualErr is non-empty but no marker
//     regex matched. Gno errored, but the wording diverges entirely
//     from gc's. Either a wording difference worth blessing with an
//     explicit `// Error:` directive, or a sign the test is firing
//     on a different defect than gc detects.
func VerifyErrorcheckMarkers(markers []InlineError, actualErr, typeCheckErr string) error {
	combined := strings.TrimSpace(actualErr + "\n" + typeCheckErr)
	if combined == "" {
		return fmt.Errorf("errorcheck: Gno accepted file but %d marker(s) expected an error",
			len(markers))
	}
	for _, m := range markers {
		if MarkerMatches(m, combined) {
			return nil
		}
	}
	return fmt.Errorf("errorcheck: Gno errored but no marker matched.\nGno output:\n%s\nmarkers:\n%s",
		indent(combined, "  "), formatMarkers(markers))
}

// PrependPkgPathIfNeeded rescues errorcheck and compile-only files
// declaring a non-main package (e.g. `package p`) so they reach
// Gno's preprocess+typecheck phase instead of bouncing on the
// realm-naming requirement. Two transforms:
//
//  1. Prepend `// PKGPATH: gno.land/p/filetest/<name>` so Gno's
//     harness stops bailing with "expected package name [main] but
//     got [p]".
//  2. Append `func main() {}` when source has no `main()` of its
//     own, so Gno's executor doesn't then bail with "name main not
//     declared" once PKGPATH gets it past the package-name check.
//
// No-op when source is already `package main` or already declares
// its own `// PKGPATH:`. The chosen pkgpath prefix is intentionally
// neutral — `gno.land/p/filetest/` — so external test drivers can
// reuse this helper without a namespace clash.
func PrependPkgPathIfNeeded(source []byte) []byte {
	if bytes.Contains(source, []byte("// PKGPATH:")) {
		return source
	}
	m := pkgDeclRe.FindSubmatch(source)
	if m == nil || string(m[1]) == "main" {
		return source
	}
	prefix := fmt.Appendf(nil, "// PKGPATH: gno.land/p/filetest/%s\n", m[1])
	out := make([]byte, 0, len(prefix)+len(source))
	out = append(out, prefix...)
	out = append(out, source...)
	if !funcMainRe.Match(source) {
		out = append(out, "\nfunc main() {}\n"...)
	}
	return out
}

// pkgDeclRe matches the Go package declaration; captures the name.
var pkgDeclRe = regexp.MustCompile(`(?m)^package (\w+)`)

// funcMainRe matches `func main()` at line start, with optional whitespace before `()`.
var funcMainRe = regexp.MustCompile(`(?m)^func main\s*\(\)`)

// IsRunnable reports whether source can be `go run` — i.e. declares
// `package main` AND has a top-level `func main()`. Files lacking
// either are compile-only by intent: gc accepts and never runs them,
// and Gno's harness must apply the PKGPATH+synthetic-main rescue (via
// [PrependPkgPathIfNeeded]) to reach preprocess+typecheck. Used by
// dispatch logic to route .go files into the compile-only path.
func IsRunnable(source []byte) bool {
	m := pkgDeclRe.FindSubmatch(source)
	if m == nil || string(m[1]) != "main" {
		return false
	}
	return funcMainRe.Match(source)
}

// reErrorLine pulls the first `:N:` line-number reference out of an
// error string. Gno errors typically format as `path:line:col: msg`
// or `[stage] path:line:col: msg`; we anchor on the first ":<digits>:"
// so prefixed-stage variations still match.
var reErrorLine = regexp.MustCompile(`:(\d+):`)

// ExtractErrorLine returns the first source-line number referenced by
// errStr, or 0 if none. Used by the errorcheck multi-pass driver to
// decide which inline `// ERROR` marker the current pass corresponds
// to. The number is in source-file coordinates; the caller is
// responsible for subtracting any prepended-line offset.
func ExtractErrorLine(errStr string) int {
	m := reErrorLine.FindStringSubmatch(errStr)
	if m == nil {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

// rePackageDecl matches a top-level package declaration line.
var rePackageDecl = regexp.MustCompile(`^package \w+`)

// NeutralizeLine neutralizes line N (1-based) so the next multi-pass
// surfaces the NEXT error, preserving the total line count so later
// error line numbers still align.
//
// A package declaration is rewritten to `package main` rather than
// commented out: the package clause is the file's one global
// dependency, so blanking it (e.g. an `// ERROR "invalid package
// name"` test on `package _`) would leave a structurally-invalid file
// and block iteration to the remaining markers. Rewriting to a valid
// `package main` lets the multi-pass continue. Every other line is an
// independent statement and is simply replaced with `//`.
//
// Returns the new source and whether the neutralized line was a
// package declaration (the caller then switches pkgPath to "main").
// Out-of-range line returns source unchanged.
func NeutralizeLine(source []byte, line int) (out []byte, wasPackage bool) {
	lines := bytes.Split(source, []byte("\n"))
	if line < 1 || line > len(lines) {
		return source, false
	}
	if rePackageDecl.Match(bytes.TrimLeft(lines[line-1], " \t")) {
		lines[line-1] = []byte("package main")
		wasPackage = true
	} else {
		lines[line-1] = []byte("//")
	}
	return bytes.Join(lines, []byte("\n")), wasPackage
}

// internalNoise reports whether a Gno error segment is an internal
// assertion rather than a real diagnostic — these surface when the
// multi-pass's neutralization pushes Gno into an inconsistent state,
// and should not be pinned in a golden if a real error for the same
// line is available.
func internalNoise(seg string) bool {
	return strings.Contains(seg, "should not happen") ||
		strings.Contains(seg, "panic:")
}

// errorForLine returns the cleaned message for the error keyed on
// gnoLine, choosing among the `; `-segments Gno reported for that line.
// Selection priority:
//  1. a marker-aligned segment (matches marker's regex) — when Gno
//     reports several errors for one line, this picks the one the gc
//     marker was about (e.g. go/types' precise "too many arguments"
//     over a vaguer preprocess "not enough arguments");
//  2. a non-internal-assertion ("should not happen") segment;
//  3. any line-keyed segment;
//  4. the first available segment (error without a position).
//
// Preprocess segments are considered before go/types within each tier.
// marker may be nil (unmarked line) to skip tier 1.
func errorForLine(errSegs, tcSegs []string, gnoLine int, marker *InlineError) string {
	key := fmt.Sprintf(":%d:", gnoLine)
	groups := [][]string{errSegs, tcSegs}

	if marker != nil {
		for _, group := range groups {
			for _, seg := range group {
				if strings.Contains(seg, key) && !internalNoise(seg) && MarkerMatches(*marker, seg) {
					return CleanErrorMessage(seg)
				}
			}
		}
	}
	for _, group := range groups {
		for _, seg := range group {
			if strings.Contains(seg, key) && !internalNoise(seg) {
				return CleanErrorMessage(seg)
			}
		}
	}
	for _, group := range groups {
		for _, seg := range group {
			if strings.Contains(seg, key) {
				return CleanErrorMessage(seg)
			}
		}
	}
	for _, group := range groups {
		if len(group) > 0 {
			return CleanErrorMessage(group[0])
		}
	}
	return ""
}

// segHasLine reports whether any segment is keyed on gnoLine (i.e.
// contains a `:<gnoLine>:` position).
func segHasLine(segs []string, gnoLine int) bool {
	key := fmt.Sprintf(":%d:", gnoLine)
	for _, s := range segs {
		if strings.Contains(s, key) {
			return true
		}
	}
	return false
}

// gnoErrSegments splits a Gno error string on "; " — Gno joins
// multiple preprocess errors into one message with that separator.
// Each segment is an independent `path:line:col: msg`. Whitespace is
// trimmed; empty segments are dropped.
func gnoErrSegments(s string) []string {
	var out []string
	for _, p := range strings.Split(s, "; ") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// CleanErrorMessage strips a leading `path:line:col:` prefix from
// errStr, leaving just the message text. Used when recording a Gno
// error in the `// GnoError:` golden block so the block doesn't pin
// volatile file paths or column numbers.
func CleanErrorMessage(errStr string) string {
	s := strings.TrimSpace(errStr)
	// Format: file:line:col: text  → split into 4 parts, take the 4th.
	parts := strings.SplitN(s, ":", 4)
	if len(parts) == 4 {
		// Sanity: parts[1] should be a number (line); parts[2] a number (col).
		if _, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
			return strings.TrimSpace(parts[3])
		}
	}
	// Fall back: try file:line: text (3 parts).
	parts = strings.SplitN(s, ":", 3)
	if len(parts) == 3 {
		if _, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
			return strings.TrimSpace(parts[2])
		}
	}
	return s
}

// FormatGnoErrorBlock serializes a per-line map of Gno error messages
// into the multi-line body of a `// GnoError:` directive. Output is
// one line per entry, in ascending line order, in the form:
//
//	line N: <Gno's error message>
//
// Empty input yields the empty string (caller writes no directive).
func FormatGnoErrorBlock(perLine map[int]string) string {
	if len(perLine) == 0 {
		return ""
	}
	lines := make([]int, 0, len(perLine))
	for n := range perLine {
		lines = append(lines, n)
	}
	sort.Ints(lines)
	var buf strings.Builder
	for _, n := range lines {
		fmt.Fprintf(&buf, "line %d: %s\n", n, perLine[n])
	}
	return strings.TrimRight(buf.String(), "\n")
}

// indent prefixes every line of s with prefix.
func indent(s, prefix string) string {
	if s == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}

// formatMarkers renders a marker list one per line for diagnostics.
func formatMarkers(ms []InlineError) string {
	var sb strings.Builder
	for _, m := range ms {
		fmt.Fprintf(&sb, "  L%d: %s\n", m.Line, strings.Join(m.Patterns, " "))
	}
	return strings.TrimRight(sb.String(), "\n")
}
