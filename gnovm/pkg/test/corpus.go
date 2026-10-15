package test

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// UsesGenerics reports whether source declares any type parameters
// (generic func or type). Gno has no generics, so such files are a
// feature gap — but Gno's failure surfaces in varied ways (preprocess
// "type parameter", or a downstream runtime "name T not declared"), so
// message matching is unreliable. Parsing the Go AST is exact. Returns
// false if source doesn't parse (e.g. a deliberately malformed
// errorcheck file) — then other detectors decide.
func UsesGenerics(source []byte) bool {
	f, err := parser.ParseFile(token.NewFileSet(), "", source, parser.SkipObjectResolution)
	if err != nil {
		return false
	}
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Type.TypeParams != nil && len(d.Type.TypeParams.List) > 0 {
				return true
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok &&
					ts.TypeParams != nil && len(ts.TypeParams.List) > 0 {
					return true
				}
			}
		}
	}
	// Constraint interfaces (`int | bool`, `~int`, `comparable`) are the
	// generics feature too, even with no type parameter in the file.
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.InterfaceType:
			for _, m := range n.Methods.List {
				switch t := m.Type.(type) {
				case *ast.BinaryExpr, *ast.UnaryExpr:
					found = true
				case *ast.Ident:
					found = found || t.Name == "comparable"
				}
			}
		case *ast.Ident:
			found = found || n.Name == "comparable"
		}
		return !found
	})
	return found
}

// HasInlineErrorMarkers reports whether source contains any
// `// ERROR ` or `// GC_ERROR ` marker (Go's errorcheck convention).
// The harness refuses such files until errorcheck mode lands.
func HasInlineErrorMarkers(source []byte) bool {
	for _, tag := range []string{"// ERROR ", "// GC_ERROR ", "// ERROR\t", "// GC_ERROR\t"} {
		if bytes.Contains(source, []byte(tag)) {
			return true
		}
	}
	return false
}

// PrependPkgPathIfNeeded rescues compile-only files
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

// funcMainRe matches `func main()` at line start. Whitespace — including
// a newline — is allowed between `func` and `main` (ancient corpus files
// like ken/*.go and fixedbugs/bug203.go write `func\nmain()`; missing
// them silently downgraded run-mode files to preprocess-only checking).
var funcMainRe = regexp.MustCompile(`(?m)^func\s+main\s*\(\)`)

// IsRunnable reports whether source can be `go run` — i.e. declares
// `package main` AND has a top-level `func main()`. Files lacking
// either are compile-only by intent: gc accepts and never runs them,
// and the harness takes them through compile mode.
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

// internalNoise reports whether a Gno error segment is an internal
// assertion rather than a real diagnostic — these surface when the
// multi-pass's neutralization pushes Gno into an inconsistent state,
// and should not be pinned in a golden if a real error for the same
// line is available.
func internalNoise(seg string) bool {
	return strings.Contains(seg, "should not happen") ||
		strings.Contains(seg, "panic:")
}

// featureGapErrors maps Gno error-message fragments to a short
// feature-gap reason. Matched against Gno's ACTUAL preprocess error
// (not the source text) so it can't false-positive on the word "go"
// in a comment or a "chan" in a string — these are the messages Gno
// emits when it hits a Go feature it doesn't implement. Ported from
// gno-go-conformance's internal/classify feature-gap cases.
var featureGapErrors = []struct {
	fragment string
	reason   string
}{
	{"channels are not permitted", "channels not supported in Gno"},
	{"goroutines are not permitted", "goroutines not supported in Gno"},
	{"imaginaries are not supported", "imaginary literals not supported in Gno"},
	{"dot imports not allowed", "dot imports not supported in Gno"},
	{"builtin identifiers cannot be shadowed", "builtin shadowing not supported in Gno"},
	{"type parameter", "generics not supported in Gno"},
	{"unknown Go type", "unsupported Go type in Gno"},
}

// langGapErrors maps a precise regex over Gno's error to a feature-gap
// reason, for language features Gno simply lacks (no complex numbers, no
// uintptr type, no `clear` builtin, no channel ops). These surface as
// "name X not declared/defined" or "... is not permitted", which a bare
// substring would over-match (e.g. a user type "complexThing"), so the
// patterns are anchored to the exact builtin/keyword. 100%-certain gaps
// only — anything ambiguous stays a // KnownIssue:.
var langGapErrors = []struct {
	re     *regexp.Regexp
	reason string
}{
	{regexp.MustCompile(`name complex(64|128)? not (declared|defined)`), "complex numbers not supported in Gno"},
	{regexp.MustCompile(`name (imag|real) not (declared|defined)`), "complex numbers not supported in Gno"},
	{regexp.MustCompile(`name uintptr not (declared|defined)`), "uintptr type not supported in Gno"},
	{regexp.MustCompile(`name clear not declared`), "clear builtin not supported in Gno"},
	{regexp.MustCompile(`channel (receive|send) is not permitted`), "channels not supported in Gno"},
	{regexp.MustCompile(`select statements are not permitted`), "channels not supported in Gno (select)"},
}

// reGuardSymbolGap captures the name in Gno's "name <Sym> not
// declared/defined" when the go/types guard rejects the same file: the
// symbol is then a Go stdlib member Gno's stdlibs lack, whatever its name.
var reGuardSymbolGap = regexp.MustCompile(`name (\w+) not (?:declared|defined)`)

// reStdlibSymbolGap matches Gno's "name <Sym> not declared/defined" for a
// known exported stdlib symbol Gno's test context doesn't provide
// (os/runtime/debug/rand/time/testing). In RUN mode `go run` already
// succeeded, so the name is a real Go symbol — a capitalized stdlib one
// is a Gno stdlib gap (route to // Unsupported:), whereas a lowercase
// user name (`xx`, `i`, `main`) would be a Gno scoping BUG and is
// deliberately NOT listed here (stays a // KnownIssue:). The symbol is
// captured so the reason is greppable per-feature: when Gno later adds
// it, grep that reason, drop the directive, and re-sync to recover the
// file. The list is explicit (not "any capitalized name") so an
// unrecognized symbol still surfaces as a bug rather than being hidden.
var reStdlibSymbolGap = regexp.MustCompile(`name (Exit|Getenv|Args|Environ|Hostname|` +
	`Caller|Callers|SetFinalizer|KeepAlive|GC|GOMAXPROCS|GOARCH|GOOS|NumCPU|NumGoroutine|` +
	`Gosched|Goexit|Stack|ReadMemStats|MemProfileRecord|MemProfileRate|MemStats|Frame|Frames|` +
	`Compiler|Breakpoint|Version|Intn|Int31|Int31n|Int63|Int63n|Float64|Float32|Seed|Perm|Shuffle|` +
	`Sleep|Tick|AllocsPerRun|Caller) not (declared|defined)`)

// UnsupportedFeatureError returns a feature-gap reason if errStr is one
// of Gno's "feature not implemented" messages (channels, goroutines,
// generics, complex, uintptr, …), or "" otherwise. Also recognizes
// "unknown import path <pkg>" — a stdlib the Gno test context lacks.
// Used to route such files to `// Unsupported:` instead of
// `// KnownIssue:`.
func UnsupportedFeatureError(errStr string) string {
	if imp := UnsupportedImport(errStr); imp != "" {
		return "unknown import path " + imp
	}
	for _, g := range featureGapErrors {
		if strings.Contains(errStr, g.fragment) {
			return g.reason
		}
	}
	for _, g := range langGapErrors {
		if g.re.MatchString(errStr) {
			return g.reason
		}
	}
	if m := reStdlibSymbolGap.FindStringSubmatch(errStr); m != nil {
		return "unsupported stdlib symbol in Gno: " + m[1]
	}
	return ""
}

// nondeterministicMarkers are substrings whose presence in a run-mode
// program's output means it can't have a stable golden: runtime panic
// dumps (addresses, signals, goroutine traces) and the harness temp
// path that `go run` echoes for tests probing their own file/line.
var nondeterministicMarkers = []string{
	"signal SIGSEGV", "signal SIGABRT", "signal SIGBUS",
	"pc=0x", "[signal ", "goroutine 1 [running]",
	"gno-filetest-go-", // harness temp dir leaked into output
}

// NondeterministicOutput reports whether out contains a marker that
// makes it non-reproducible (panic addresses, goroutine dumps, the
// harness temp path). Such run-mode files are routed to
// `// Unsupported:` since neither side's output can be pinned.
func NondeterministicOutput(out string) string {
	for _, m := range nondeterministicMarkers {
		if strings.Contains(out, m) {
			return "non-deterministic runtime output (" + m + ")"
		}
	}
	return ""
}

// reUnknownImport matches an "unknown import path <path>" error,
// capturing the import path. Handles both Gno's own preprocess form
// (`unknown import path reflect`) and the go/types guard's form
// (`could not import reflect (unknown import path "reflect")`), so the
// optional quote and trailing `)` are excluded from the capture.
var reUnknownImport = regexp.MustCompile(`unknown import path "?([^\s;")]+)`)

// UnsupportedImport returns the import path from an "unknown import
// path <path>" error in errStr, or "" if none. Used to route
// errorcheck files that fail on a Gno-unsupported stdlib import
// (unsafe, syscall, net/http, …) to `// Unsupported:` rather than
// `// KnownIssue:` — Gno can't process the file at all, so it's a
// feature gap, not an over-strict bug.
func UnsupportedImport(errStr string) string {
	m := reUnknownImport.FindStringSubmatch(errStr)
	if m == nil {
		return ""
	}
	return m[1]
}

// looksLikePackagePath reports whether imp is a normal Go import path
// (a missing stdlib like "reflect" or "encoding/xml") rather than an
// invalid-import-path syntax test case ("/foo", "c:/foo", empty,
// control chars). The go/types guard emits "unknown import path" for
// both; only the former is a genuine missing-package feature gap worth
// routing to `// Unsupported:`. The latter is a real errorcheck case
// (e.g. import6.go) whose golden must be kept.
func looksLikePackagePath(imp string) bool {
	if imp == "" || strings.HasPrefix(imp, "/") || strings.HasPrefix(imp, ".") {
		return false
	}
	for _, r := range imp {
		if r < 0x20 || r == '\\' || r == ':' || r == '"' || r == ' ' {
			return false
		}
	}
	return true
}

// CorpusDirective returns the gc test directive word on source's first
// line ("compile", "run", "errorcheck", …) with any flags stripped, or
// "" if the first line isn't a recognized directive. Used to route a
// `// compile` file to compile-mode regardless of whether it happens to
// be runnable (gc compiles but never runs it).
func CorpusDirective(source []byte) string {
	first := source
	if before, _, ok := bytes.Cut(source, []byte{'\n'}); ok {
		first = before
	}
	s := strings.TrimSpace(string(first))
	if !strings.HasPrefix(s, "//") {
		return ""
	}
	s = strings.TrimSpace(s[2:])
	if i := strings.IndexAny(s, " \t"); i >= 0 {
		s = s[:i]
	}
	switch s {
	case "compile", "run", "errorcheck", "build",
		"compiledir", "rundir", "errorcheckdir", "runoutput", "errorcheckoutput", "asmcheck":
		return s
	}
	return ""
}

// PerLineErrors maps source line → cleaned message for each `; `-segment
// in errStr that carries a position, adjusting Gno's line numbers by
// prependedLines (the PKGPATH-rescue offset). Used by compile-mode to
// pin Gno / go/types errors per line.
func PerLineErrors(errStr string, prependedLines int) map[int]string {
	out := make(map[int]string)
	for _, seg := range gnoErrSegments(errStr) {
		gnoLine := ExtractErrorLine(seg)
		if gnoLine == 0 || internalNoise(seg) {
			continue
		}
		src := gnoLine - prependedLines
		if _, ok := out[src]; !ok {
			out[src] = CleanErrorMessage(seg)
		}
	}
	return out
}

// gnoErrSegments splits a Gno error string on "; " — Gno joins
// multiple preprocess errors into one message with that separator.
// Each segment is an independent `path:line:col: msg`. Whitespace is
// trimmed; empty segments are dropped.
func gnoErrSegments(s string) []string {
	var out []string
	for p := range strings.SplitSeq(s, "; ") {
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

// FormatLineOnlyBlock serializes the keys of a per-line error map as
// `line N` entries, dropping the messages. Used for the Go-side blocks
// (`// GoTypeCheckError:`, `// GoBuildError:`), whose wording changes
// with the Go release while the set of flagged lines does not.
func FormatLineOnlyBlock(perLine map[int]string) string {
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
		fmt.Fprintf(&buf, "line %d\n", n)
	}
	return strings.TrimRight(buf.String(), "\n")
}

var reLineEntry = regexp.MustCompile(`^line (\d+)(?::\s*(.*))?$`)

// ParseLineBlock reads a `// GnoError:`-style block body back into a
// per-line map; entries without a message (FormatLineOnlyBlock) map to "".
func ParseLineBlock(block string) map[int]string {
	out := make(map[int]string)
	for l := range strings.SplitSeq(block, "\n") {
		if m := reLineEntry.FindStringSubmatch(strings.TrimSpace(l)); m != nil {
			n, _ := strconv.Atoi(m[1])
			out[n] = m[2]
		}
	}
	return out
}

// unsupportedTags maps the harness's free-text skip reasons to a fixed
// tag, first match wins. The tag set is shared with gno-go-conformance's
// compat detector so both count the same buckets.
var unsupportedTags = []struct{ tag, needle string }{
	{"generics", "generics not supported"},
	{"channel", "channels not supported"},
	{"goroutine", "goroutines not supported"},
	{"complex-types", "complex numbers not supported"},
	{"imaginary-literal", "imaginary literals"},
	{"uintptr", "uintptr type not supported"},
	{"builtin-shadow", "builtin shadowing"},
	{"clear-builtin", "clear builtin"},
	{"dot-import", "dot import"},
	{"unsafe", "unknown import path unsafe"},
	{"reflect", "unknown import path reflect"},
	{"import", "unknown import path"},
	{"stdlib-symbol", "unsupported stdlib symbol"},
	{"gc-diagnostic", "gc optimization-diagnostic"},
	{"gc-lang-pin", "version-gating"},
	{"gc-specific", "only gc-specific"},
	{"gc-specific", "gc runtime-package"},
	{"gc-specific", "gc pragma-placement"},
	{"gc-specific", "gc liveness"},
	{"gc-specific", "gc's "},
	{"gc-specific", "gc-internal"},
	{"gc-specific", "//go:"},
	{"long", "in the interpreter"},
	{"nondeterministic", "non-deterministic"},
	{"nondeterministic", "nondeterministic"},
	{"marker-uncheckable", "marker on a non-code line"},
	{"marker-uncheckable", "not checkable"},
	{"runtime-vs-compile", "at runtime"},
}

// TagUnsupportedReason returns `<tag>: <reason>`, leaving a reason that
// already carries a known tag unchanged.
func TagUnsupportedReason(reason string) string {
	reason = strings.TrimSpace(reason)
	for _, t := range unsupportedTags {
		if strings.HasPrefix(reason, t.tag+": ") {
			return reason
		}
	}
	if strings.HasPrefix(reason, "other: ") {
		return reason
	}
	for _, t := range unsupportedTags {
		if strings.Contains(reason, t.needle) {
			return t.tag + ": " + reason
		}
	}
	return "other: " + reason
}
