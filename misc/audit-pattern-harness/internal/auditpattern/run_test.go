package auditpattern

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// pkgDir is the directory containing this test file, resolved at init time via
// runtime.Caller so that path computations work regardless of working directory.
var pkgDir = func() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Dir(f)
}()

func harnessRoot() string { return filepath.Clean(filepath.Join(pkgDir, "..", "..")) }
func repoRoot() string    { return filepath.Clean(filepath.Join(harnessRoot(), "..", "..")) }
func fixturesDir(name string) string {
	return filepath.Join(harnessRoot(), "fixtures", name)
}
func expectedFile(name string) string {
	return filepath.Join(harnessRoot(), "expected", name+".yaml")
}

func TestLoadRecordResolvesFixturePaths(t *testing.T) {
	rec, err := LoadRecord(expectedFile("current-guard"))
	if err != nil {
		t.Fatal(err)
	}
	if rec.ID != "current-guard" {
		t.Fatalf("unexpected id %q", rec.ID)
	}
	if len(rec.Fixtures) != 2 {
		t.Fatalf("expected 2 fixtures, got %d", len(rec.Fixtures))
	}
	if !filepath.IsAbs(rec.Fixtures[0].Path) {
		t.Fatalf("fixture path is not absolute: %s", rec.Fixtures[0].Path)
	}
}

func TestCurrentGuardRule(t *testing.T) {
	base := fixturesDir("current-guard")

	hits, err := RunRule("current_guard", filepath.Join(base, "vulnerable"))
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].File != "admin.gno" || hits[0].Line != 7 {
		t.Fatalf("unexpected hit: %+v", hits[0])
	}

	hits, err = RunRule("current_guard", filepath.Join(base, "fixed"))
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Fatalf("expected no hits, got %d", len(hits))
	}
}

func TestRenderMarkdownEscapeRule(t *testing.T) {
	base := fixturesDir("render-markdown")

	hits, err := RunRule("render_markdown_escape", filepath.Join(base, "vulnerable"))
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].File != "echo.gno" {
		t.Fatalf("unexpected hit: %+v", hits[0])
	}

	hits, err = RunRule("render_markdown_escape", filepath.Join(base, "fixed"))
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Fatalf("expected no hits, got %d", len(hits))
	}
}

func TestPaymentUserCallRule(t *testing.T) {
	assertRuleCounts(t, "payment_user_call", "payment-user-call", 1, 0)
}

func TestOriginCallerAuthRule(t *testing.T) {
	assertRuleCounts(t, "origin_caller_auth", "origin-caller-auth", 1, 0)
}

func TestOriginCallerBenignRead(t *testing.T) {
	dir := t.TempDir()
	src := "package x\n\n" +
		"import \"chain/runtime/unsafe\"\n\n" +
		"func Log() {\n" +
		"\temit(\"actor\", unsafe.OriginCaller().String())\n" +
		"}\n"
	if err := os.WriteFile(filepath.Join(dir, "a.gno"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	hits, err := RunRule("origin_caller_auth", dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Fatalf("benign OriginCaller read flagged as auth: %+v", hits)
	}
}

func TestCallbackParamRule(t *testing.T) {
	assertRuleCounts(t, "callback_param", "callback-param", 1, 0)
}

func TestInterfaceRealmParamRule(t *testing.T) {
	assertRuleCounts(t, "interface_realm_param", "interface-realm-param", 1, 0)
}

func TestExportedPointerLeakRule(t *testing.T) {
	assertRuleCounts(t, "exported_pointer_leak", "exported-pointer-leak", 2, 0)
}

func TestExportedPointerLeakIgnoresFreshConstructor(t *testing.T) {
	dir := t.TempDir()
	src := "package x\n\n" +
		"type Vault struct{ B int }\n\n" +
		"func NewVault() *Vault {\n" +
		"\treturn &Vault{}\n" +
		"}\n"
	if err := os.WriteFile(filepath.Join(dir, "a.gno"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	hits, err := RunRule("exported_pointer_leak", dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Fatalf("fresh constructor flagged as pointer leak: %+v", hits)
	}
}

// TestRuleNormalizesFormatting ensures the gofmt pre-step lets matchers catch
// badly-formatted source that would otherwise slip past spacing-sensitive
// checks (e.g. "func GetVault()*Vault{" instead of "func GetVault() *Vault {").
func TestRuleNormalizesFormatting(t *testing.T) {
	dir := t.TempDir()
	src := "package vault\n\n" +
		"type Vault struct{ items []string }\n\n" +
		"func GetVault()*Vault{\n\treturn &Vault{}\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "vault.gno"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	hits, err := RunRule("exported_pointer_leak", dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit after formatting normalization, got %d: %+v", len(hits), hits)
	}
}

// TestBraceInStringNoFalsePositive ensures a "}" inside a string literal does
// not flip brace-depth tracking and flag a correctly guarded function.
func TestBraceInStringNoFalsePositive(t *testing.T) {
	dir := t.TempDir()
	src := "package x\n\n" +
		"func F(cur realm) {\n" +
		"\tif !cur.IsCurrent() {\n" +
		"\t\tpanic(\"no\")\n" +
		"\t}\n" +
		"\tmsg := \"}\"\n" +
		"\t_ = cur.Previous()\n" +
		"\t_ = msg\n" +
		"}\n"
	if err := os.WriteFile(filepath.Join(dir, "a.gno"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	hits, err := RunRule("current_guard", dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Fatalf("guarded function flagged due to brace in string: %+v", hits)
	}
}

func TestRenderMapIterationRule(t *testing.T) {
	assertRuleCounts(t, "render_map_iteration", "render-map-iteration", 1, 0)
}

func TestRunWithFakeGNO(t *testing.T) {
	tmp := t.TempDir()
	gno := filepath.Join(tmp, "gno")
	if err := os.WriteFile(gno, []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	rec, err := LoadRecord(expectedFile("current-guard"))
	if err != nil {
		t.Fatal(err)
	}

	report := Run(context.Background(), rec, Options{GNOBin: gno})
	if !report.OK {
		t.Fatalf("expected report to pass: %+v", report)
	}
	if len(report.Fixtures) != 2 {
		t.Fatalf("expected 2 fixture results, got %d", len(report.Fixtures))
	}
}

func TestAgentPatternContract(t *testing.T) {
	specFiles := []string{
		filepath.Join(harnessRoot(), "README.md"),
		filepath.Join(repoRoot(), "docs", "resources", "gno-security-guide.md"),
		filepath.Join(repoRoot(), "docs", "resources", "effective-gno.md"),
		filepath.Join(repoRoot(), "docs", "resources", "gno-data-structures.md"),
		filepath.Join(repoRoot(), "docs", "resources", "community-packages.md"),
	}
	requiredTerms := map[string][]string{
		"callback-param":        {"callback", "caller-supplied"},
		"current-guard":         {"cur.Previous()", "cur.IsCurrent()"},
		"exported-pointer-leak": {"exported pointer", "mutable state"},
		"interface-realm-param": {"interface", "cur realm"},
		"origin-caller-auth":    {"OriginCaller()", "authorization"},
		"payment-user-call":     {"OriginSend()", "IsUserCall()"},
		"render-map-iteration":  {"Render", "map iteration"},
		"render-markdown":       {"Render(path)", "markdown/sanitize"},
	}
	// Every vulnerable hit must contain the rule's detection signal. Counting
	// hits alone lets a rule be rewritten to flag a coincidental line (e.g. an
	// import) while the suite stays green; this ties the hit to the construct
	// the rule is supposed to detect.
	wantHitContains := map[string]string{
		"callback-param":        "func(",
		"current-guard":         ".Previous()",
		"exported-pointer-leak": "*",
		"interface-realm-param": "realm",
		"origin-caller-auth":    "OriginCaller()",
		"payment-user-call":     "OriginSend()",
		"render-map-iteration":  "range ",
		"render-markdown":       "path",
	}

	corpus := readSpecCorpus(t, specFiles)
	records := loadAllRecords(t, harnessRoot())
	if len(records) < len(requiredTerms) {
		t.Fatalf("expected at least %d records, got %d", len(requiredTerms), len(records))
	}

	for _, rec := range records {
		t.Run(rec.ID, func(t *testing.T) {
			for _, term := range requiredTerms[rec.ID] {
				if !strings.Contains(corpus, term) {
					t.Fatalf("spec corpus does not mention %q for %s", term, rec.ID)
				}
			}

			fixtures := map[string]Fixture{}
			for _, fixture := range rec.Fixtures {
				fixtures[fixture.Name] = fixture
			}
			vulnerable, ok := fixtures["vulnerable"]
			if !ok {
				t.Fatalf("missing vulnerable fixture")
			}
			fixed, ok := fixtures["fixed"]
			if !ok {
				t.Fatalf("missing fixed fixture")
			}
			if vulnerable.WantPatternHits <= 0 {
				t.Fatalf("vulnerable fixture must expect at least one pattern hit")
			}
			if fixed.WantPatternHits != 0 {
				t.Fatalf("fixed fixture must expect zero pattern hits")
			}

			vulnerableHits, err := RunRule(rec.Rule, vulnerable.Path)
			if err != nil {
				t.Fatal(err)
			}
			if len(vulnerableHits) != vulnerable.WantPatternHits {
				t.Fatalf("vulnerable hits: got %d, want %d: %+v", len(vulnerableHits), vulnerable.WantPatternHits, vulnerableHits)
			}
			if marker := wantHitContains[rec.ID]; marker != "" {
				for _, h := range vulnerableHits {
					if !strings.Contains(h.Text, marker) {
						t.Fatalf("vulnerable hit %q does not contain rule signal %q; rule may be matching a coincidental line", h.Text, marker)
					}
				}
			}

			fixedHits, err := RunRule(rec.Rule, fixed.Path)
			if err != nil {
				t.Fatal(err)
			}
			if len(fixedHits) != 0 {
				t.Fatalf("fixed fixture still has pattern hits: %+v", fixedHits)
			}
		})
	}
}

func TestAgentPatternContractWithGNO(t *testing.T) {
	gnoBin := os.Getenv("GNO_BIN")
	if gnoBin == "" {
		var err error
		gnoBin, err = exec.LookPath("gno")
		if err != nil {
			t.Skip("set GNO_BIN or install gno in PATH to compile-check all agent contract fixtures")
		}
	}

	for _, rec := range loadAllRecords(t, harnessRoot()) {
		t.Run(rec.ID, func(t *testing.T) {
			report := Run(context.Background(), rec, Options{GNOBin: gnoBin})
			if !report.OK {
				t.Fatalf("agent contract failed with %s: %+v", gnoBin, report)
			}
		})
	}
}

func assertRuleCounts(t *testing.T, rule, fixture string, vulnerable, fixed int) {
	t.Helper()
	base := fixturesDir(fixture)

	hits, err := RunRule(rule, filepath.Join(base, "vulnerable"))
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != vulnerable {
		t.Fatalf("expected %d vulnerable hits, got %d: %+v", vulnerable, len(hits), hits)
	}

	hits, err = RunRule(rule, filepath.Join(base, "fixed"))
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != fixed {
		t.Fatalf("expected %d fixed hits, got %d: %+v", fixed, len(hits), hits)
	}
}

func readSpecCorpus(t *testing.T, paths []string) string {
	t.Helper()

	var corpus strings.Builder
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read spec file %s: %v", path, err)
		}
		if len(strings.TrimSpace(string(data))) == 0 {
			t.Fatalf("spec file %s is empty", path)
		}
		corpus.Write(data)
		corpus.WriteByte('\n')
	}
	return corpus.String()
}

func loadAllRecords(t *testing.T, harnessRoot string) []Record {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(harnessRoot, "expected", "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		t.Fatalf("no expected records found")
	}

	seen := map[string]bool{}
	records := make([]Record, 0, len(paths))
	for _, path := range paths {
		rec, err := LoadRecord(path)
		if err != nil {
			t.Fatal(err)
		}
		if seen[rec.ID] {
			t.Fatalf("duplicate record id %q", rec.ID)
		}
		seen[rec.ID] = true
		records = append(records, rec)
	}
	return records
}
