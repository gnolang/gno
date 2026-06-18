package auditloop

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
