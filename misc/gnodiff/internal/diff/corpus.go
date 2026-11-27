package diff

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Program is one seed, or one fuzz input, compared on both toolchains.
type Program struct {
	Name   string
	Source string
	// Known is the reason from a `// gnodiff:known ...` comment. A divergence
	// on a known program is reported and does not fail the run.
	Known string
}

// CaseResult is one program after both sides have run.
type CaseResult struct {
	Name      string `json:"name"`
	Known     string `json:"known,omitempty"`
	Match     bool   `json:"match"`
	Reason    string `json:"reason,omitempty"`
	Go        Result `json:"go"`
	Gno       Result `json:"gno"`
	Minimized string `json:"minimized,omitempty"`
}

// Report is the full corpus run.
type Report struct {
	Total     int          `json:"total"`
	Matched   int          `json:"matched"`
	Known     int          `json:"known"`
	Divergent int          `json:"divergent"`
	Internal  int          `json:"internal"`
	Cases     []CaseResult `json:"cases"`
}

// Failed is true when an unexpected divergence or a runner failure was seen.
// Annotated known divergences do not fail the run.
func (r Report) Failed() bool {
	return r.Divergent > 0 || r.Internal > 0
}

// RunConfig is the corpus driver.
type RunConfig struct {
	Options        Options
	Minimize       bool
	MinimizeBudget int
	OnCase         func(CaseResult)
}

// LoadDir reads corpus/<name>/prog.src, one program per directory.
// The seeds are not .go files: `go test` would vet them, and several
// interesting cases (integer-to-string, variable slice bounds) are legal
// programs that vet rejects.
func LoadDir(dir string) ([]Program, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read corpus %s: %w", dir, err)
	}
	var out []Program
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name(), "prog.src")
		body, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		src := string(body)
		out = append(out, Program{
			Name:   entry.Name(),
			Source: src,
			Known:  knownReason(src),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	if len(out) == 0 {
		return nil, fmt.Errorf("no programs in %s (expected <name>/prog.src)", dir)
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

func knownReason(src string) string {
	const prefix = "// gnodiff:known"
	for line := range strings.SplitSeq(src, "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, prefix); ok {
			reason := strings.TrimSpace(after)
			if reason == "" {
				return "annotated"
			}
			return reason
		}
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
		switch {
		case item.Go.Kind == KindInternal || item.Gno.Kind == KindInternal:
			rep.Internal++
		case item.Match:
			rep.Matched++
		case prog.Known != "":
			rep.Known++
		default:
			rep.Divergent++
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
	match, reason := Compare(goRes, gnoRes)
	item := CaseResult{
		Name:   prog.Name,
		Known:  prog.Known,
		Match:  match,
		Reason: reason,
		Go:     goRes,
		Gno:    gnoRes,
	}
	if match || prog.Known != "" || !cfg.Minimize {
		return item
	}
	item.Minimized = Minimize(prog.Source, cfg.MinimizeBudget, func(src string) bool {
		left := RunGo(context.Background(), src, cfg.Options)
		right := RunGno(src, cfg.Options)
		ok, _ := Compare(left, right)
		return !ok
	})
	return item
}

// Text renders a human report. Known divergences are listed separately from
// unexpected ones.
func (r Report) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "gnodiff: %d programs, %d matched, %d known, %d unexpected, %d runner failures\n",
		r.Total, r.Matched, r.Known, r.Divergent, r.Internal)
	if r.Divergent == 0 && r.Internal == 0 && r.Known == 0 {
		b.WriteString("no divergences.\n")
		return b.String()
	}
	b.WriteByte('\n')
	for _, item := range r.Cases {
		if item.Match && item.Known == "" {
			continue
		}
		label := "DIVERGENCE"
		switch {
		case item.Go.Kind == KindInternal || item.Gno.Kind == KindInternal:
			label = "RUNNER"
		case item.Match && item.Known != "":
			label = "KNOWN-AGREES"
		case item.Known != "":
			label = "KNOWN"
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
	reproDir := filepath.Join(dir, "repro")
	for _, item := range r.Cases {
		if item.Minimized == "" {
			continue
		}
		if err := os.MkdirAll(reproDir, 0o755); err != nil {
			return err
		}
		header := fmt.Sprintf("// gnodiff repro %s: %s\n// go: %s\n// gno: %s\n",
			item.Name, item.Reason, oneLine(item.Go.Summary()), oneLine(item.Gno.Summary()))
		path := filepath.Join(reproDir, item.Name+".go")
		if err := os.WriteFile(path, []byte(header+item.Minimized), 0o644); err != nil {
			return err
		}
	}
	return nil
}
