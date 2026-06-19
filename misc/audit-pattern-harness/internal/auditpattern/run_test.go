package auditpattern

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRecordResolvesFixturePaths(t *testing.T) {
	rec, err := LoadRecord(filepath.Join("..", "..", "expected", "current-guard.yaml"))
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
	base := filepath.Join("..", "..", "fixtures", "current-guard")

	hits, err := RunRule("current_guard", filepath.Join(base, "vulnerable"))
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].File != "admin.gno" || hits[0].Line != 6 {
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
	base := filepath.Join("..", "..", "fixtures", "render-markdown")

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

func TestCallbackParamRule(t *testing.T) {
	assertRuleCounts(t, "callback_param", "callback-param", 1, 0)
}

func TestInterfaceRealmParamRule(t *testing.T) {
	assertRuleCounts(t, "interface_realm_param", "interface-realm-param", 1, 0)
}

func TestExportedPointerLeakRule(t *testing.T) {
	assertRuleCounts(t, "exported_pointer_leak", "exported-pointer-leak", 2, 0)
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

	rec, err := LoadRecord(filepath.Join("..", "..", "expected", "current-guard.yaml"))
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

func assertRuleCounts(t *testing.T, rule, fixture string, vulnerable, fixed int) {
	t.Helper()
	base := filepath.Join("..", "..", "fixtures", fixture)

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
