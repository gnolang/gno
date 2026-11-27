package diff

import (
	"context"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/gnolang/gno/misc/gnodiff/internal/gen"
)

// Program is one seed, or one fuzz input, compared on both toolchains.
type Program struct {
	Name   string
	Source string
	// Known is the reason from a `// gnodiff:known ...` comment. A divergence
	// on a known program is reported and does not fail the run, unless an
	// expect line is present and the output moved.
	Known     string
	ExpectGo  string
	ExpectGno string
}

// CaseResult is one program after both sides have run.
type CaseResult struct {
	Name      string `json:"name"`
	Class     Class  `json:"class"`
	Known     string `json:"known,omitempty"`
	Reason    string `json:"reason,omitempty"`
	Go        Result `json:"go"`
	Gno       Result `json:"gno"`
	Lint      Result `json:"lint,omitempty"`
	Minimized string `json:"minimized,omitempty"`
	Source    string `json:"source,omitempty"`
}

// Report is the full corpus run.
type Report struct {
	Total         int          `json:"total"`
	Exact         int          `json:"exact"`
	Known         int          `json:"known"`
	KnownAgrees   int          `json:"knownAgrees"`
	PanicWording  int          `json:"panicWording"`
	RejectWording int          `json:"rejectWording"`
	LintRun       int          `json:"lintRun"`
	Real          int          `json:"real"`
	Internal      int          `json:"internal"`
	Cases         []CaseResult `json:"cases"`
}

// Failed is true when a real divergence or a runner failure was seen.
// Known differences, panic-wording differences, and "run accepts what lint
// rejects" do not fail the run.
func (r Report) Failed() bool {
	return r.Real > 0 || r.Internal > 0
}

// RunConfig is the corpus driver.
type RunConfig struct {
	Options        Options
	Minimize       bool
	MinimizeBudget int
	OnCase         func(CaseResult)
}

// LoadDir reads every *.src file under dir. corpus/<name>/prog.src keeps the
// short name <name>. corpus/<category>/<id>.src is named <category>/<id>.
// Seeds are not .go files: `go test` would vet them, and several interesting
// cases are legal programs that vet rejects.
func LoadDir(dir string) ([]Program, error) {
	var out []Program
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".src") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		name := strings.TrimSuffix(rel, ".src")
		name = strings.TrimSuffix(name, string(filepath.Separator)+"prog")
		src := string(body)
		out = append(out, Program{
			Name:      name,
			Source:    src,
			Known:     directive(src, "known"),
			ExpectGo:  directive(src, "expect-go"),
			ExpectGno: directive(src, "expect-gno"),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read corpus %s: %w", dir, err)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	if len(out) == 0 {
		return nil, fmt.Errorf("no programs in %s (expected *.src)", dir)
	}
	return out, nil
}

// CorpusDir finds the seed directory from the working directory or from
// this source file, so `go run .` and `go test` agree.
func CorpusDir() string {
	candidates := []string{"corpus"}
	if _, file, _, ok := runtime.Caller(0); ok {
		base := filepath.Dir(file)
		candidates = append(candidates,
			filepath.Join(base, "corpus"),
			filepath.Join(base, "..", "..", "corpus"),
		)
	}
	for _, candidate := range candidates {
		st, err := os.Stat(candidate)
		if err == nil && st.IsDir() {
			return candidate
		}
	}
	return "corpus"
}

// LoadCorpus reads the hand-written seeds and appends the enumerated catalog,
// which is built in memory rather than checked in.
func LoadCorpus() ([]Program, error) {
	hand, err := LoadDir(CorpusDir())
	if err != nil {
		return nil, err
	}
	return WithCatalog(hand), nil
}

// WithCatalog appends RawCatalog programs. A hand seed with the same name
// wins, so a checked-in annotation is not overwritten.
func WithCatalog(hand []Program) []Program {
	seen := map[string]bool{}
	for _, prog := range hand {
		seen[prog.Name] = true
	}
	for _, seed := range gen.RawCatalog() {
		name := seed.Category + "/" + seed.Name
		if seen[name] {
			continue
		}
		src := seed.Source
		if formatted, err := format.Source([]byte(src)); err == nil {
			src = string(formatted)
		}
		hand = append(hand, Program{
			Name:      name,
			Source:    src,
			Known:     directive(src, "known"),
			ExpectGo:  directive(src, "expect-go"),
			ExpectGno: directive(src, "expect-gno"),
		})
		seen[name] = true
	}
	sort.Slice(hand, func(i, j int) bool { return hand[i].Name < hand[j].Name })
	return hand
}

func directive(src, key string) string {
	prefix := "// gnodiff:" + key
	for line := range strings.SplitSeq(src, "\n") {
		line = strings.TrimSpace(line)
		after, ok := strings.CutPrefix(line, prefix)
		if !ok {
			continue
		}
		// // gnodiff:go must not match // gnodiff:gno. Require end or space.
		if after != "" && !strings.HasPrefix(after, " ") && !strings.HasPrefix(after, "\t") {
			continue
		}
		reason := strings.TrimSpace(after)
		if reason == "" {
			return "annotated"
		}
		return reason
	}
	return ""
}

// RunCorpus executes every program on both toolchains.
func RunCorpus(progs []Program, cfg RunConfig) Report {
	if cfg.MinimizeBudget <= 0 {
		cfg.MinimizeBudget = 30
	}
	opt, err := cfg.Options.resolve()
	if err != nil {
		rep := Report{Total: len(progs), Internal: len(progs)}
		for _, prog := range progs {
			rep.Cases = append(rep.Cases, CaseResult{
				Name:   prog.Name,
				Class:  ClassReal,
				Known:  prog.Known,
				Reason: "runner failed",
				Go:     Result{Kind: KindInternal, Detail: err.Error()},
				Gno:    Result{Kind: KindInternal, Detail: err.Error()},
			})
		}
		return rep
	}
	cfg.Options = opt

	var rep Report
	rep.Total = len(progs)
	for _, prog := range progs {
		item := runOne(prog, cfg)
		switch item.Class {
		case ClassExact:
			rep.Exact++
		case ClassKnown:
			rep.Known++
		case ClassKnownAgrees:
			rep.KnownAgrees++
		case ClassPanicWording:
			rep.PanicWording++
		case ClassRejectWording:
			rep.RejectWording++
		case ClassLintRun:
			rep.LintRun++
		default:
			if item.Go.Kind == KindInternal || item.Gno.Kind == KindInternal || item.Lint.Kind == KindInternal ||
				item.Reason == "runner failed" || item.Reason == "lint runner failed" {
				rep.Internal++
			} else {
				rep.Real++
			}
		}
		rep.Cases = append(rep.Cases, item)
		if cfg.OnCase != nil {
			cfg.OnCase(item)
		}
	}
	return rep
}

func runOne(prog Program, cfg RunConfig) CaseResult {
	goRes := RunGo(context.Background(), prog.Source, cfg.Options)
	gnoRes := RunGno(prog.Source, cfg.Options)
	lintRes := RunLint(prog.Source, cfg.Options)
	class, reason := Classify(prog, goRes, gnoRes, lintRes)
	item := CaseResult{
		Name:   prog.Name,
		Class:  class,
		Known:  prog.Known,
		Reason: reason,
		Go:     goRes,
		Gno:    gnoRes,
		Lint:   lintRes,
	}
	if class != ClassExact && class != ClassKnownAgrees {
		item.Source = prog.Source
	}
	if class != ClassReal || !cfg.Minimize {
		return item
	}
	if goRes.Kind == KindInternal || gnoRes.Kind == KindInternal {
		return item
	}
	item.Minimized = Minimize(prog.Source, cfg.MinimizeBudget, func(src string) bool {
		left := RunGo(context.Background(), src, cfg.Options)
		right := RunGno(src, cfg.Options)
		lint := RunLint(src, cfg.Options)
		cl, _ := Classify(Program{Source: src}, left, right, lint)
		return cl == ClassReal && left.Kind != KindInternal && right.Kind != KindInternal && lint.Kind != KindInternal
	})
	return item
}

// Text renders a human report. Known divergences are listed separately from
// unexpected ones.
func (r Report) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "gnodiff: %d programs, %d exact, %d known, %d known-agrees, %d panic-wording, %d reject-wording, %d lint-run, %d real, %d runner failures\n",
		r.Total, r.Exact, r.Known, r.KnownAgrees, r.PanicWording, r.RejectWording, r.LintRun, r.Real, r.Internal)
	if r.Real == 0 && r.Internal == 0 && r.Known == 0 && r.PanicWording == 0 {
		b.WriteString("no divergences.\n")
		return b.String()
	}
	b.WriteByte('\n')
	for _, item := range r.Cases {
		if item.Class == ClassExact {
			continue
		}
		label := "REAL"
		switch item.Class {
		case ClassKnown:
			label = "KNOWN"
		case ClassKnownAgrees:
			label = "KNOWN-AGREES"
		case ClassPanicWording:
			label = "WORDING"
		case ClassRejectWording:
			label = "REJECT-WORDING"
		case ClassLintRun:
			label = "LINT-RUN"
		}
		fmt.Fprintf(&b, "%s %s", label, item.Name)
		if item.Reason != "" {
			fmt.Fprintf(&b, " (%s)", item.Reason)
		}
		b.WriteByte('\n')
		if item.Known != "" {
			fmt.Fprintf(&b, "  note: %s\n", item.Known)
		}
		fmt.Fprintf(&b, "  go:  %s\n", indentSummary(item.Go.Summary()))
		fmt.Fprintf(&b, "  gno: %s\n", indentSummary(item.Gno.Summary()))
		if item.Lint.Kind != "" {
			fmt.Fprintf(&b, "  lint: %s\n", indentSummary(item.Lint.Summary()))
		}
		if item.Gno.Gas > 0 {
			fmt.Fprintf(&b, "  gno gas: %d (recorded, not compared)\n", item.Gno.Gas)
		}
		if item.Minimized != "" {
			b.WriteString("  minimized repro saved\n")
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func indentSummary(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) == 1 {
		return lines[0]
	}
	return lines[0] + "\n    " + strings.Join(lines[1:], "\n    ")
}

// Write persists the text report, a JSON report, and minimized reproducers.
func (r Report) Write(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "report.txt"), []byte(r.Text()), 0o644); err != nil {
		return err
	}
	body, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	if err := os.WriteFile(filepath.Join(dir, "report.json"), body, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "panic-wording.txt"), []byte(r.wordingText()), 0o644); err != nil {
		return err
	}
	reproDir := filepath.Join(dir, "repro")
	for _, item := range r.Cases {
		body := item.Minimized
		if body == "" {
			continue
		}
		if err := os.MkdirAll(reproDir, 0o755); err != nil {
			return err
		}
		header := fmt.Sprintf("// gnodiff repro %s (%s): %s\n// go: %s\n// gno: %s\n",
			item.Name, item.Class, item.Reason, oneLine(item.Go.Summary()), oneLine(item.Gno.Summary()))
		safe := strings.NewReplacer("/", "__", "\\", "__").Replace(item.Name)
		path := filepath.Join(reproDir, safe+".go")
		if err := os.WriteFile(path, []byte(header+body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// wordingText lists every distinct panic-message pair with one repro each,
// so the texts can be filed as compatibility issues.
func (r Report) wordingText() string {
	type pair struct {
		goMsg  string
		gnoMsg string
		names  []string
		src    string
	}
	var pairs []pair
	index := map[string]int{}
	for _, item := range r.Cases {
		if item.Class != ClassPanicWording {
			continue
		}
		key := cleanPanic(item.Go.Panic) + "\n" + cleanPanic(item.Gno.Panic)
		i, ok := index[key]
		if !ok {
			index[key] = len(pairs)
			pairs = append(pairs, pair{
				goMsg:  cleanPanic(item.Go.Panic),
				gnoMsg: cleanPanic(item.Gno.Panic),
				src:    item.Source,
			})
			i = len(pairs) - 1
		}
		pairs[i].names = append(pairs[i].names, item.Name)
		if len(item.Source) < len(pairs[i].src) {
			pairs[i].src = item.Source
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# panic wording differences (%d programs, %d distinct pairs)\n\n", r.PanicWording, len(pairs))
	if len(pairs) == 0 {
		b.WriteString("none.\n")
		return b.String()
	}
	for n, p := range pairs {
		fmt.Fprintf(&b, "## %d\ngo:  %s\ngno: %s\nprograms: %s\n", n+1, p.goMsg, p.gnoMsg, strings.Join(p.names, ", "))
		b.WriteString("--- repro ---\n")
		b.WriteString(p.src)
		if !strings.HasSuffix(p.src, "\n") {
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	return b.String()
}
