package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestRenderJSON(t *testing.T) {
	r := Report{
		GeneratedAt: time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
		WindowDays:  30,
		Sections: []Section{
			{Name: "Hot", Entries: []Entry{
				{Repo: "gnolang/gno", Number: 100, Kind: KindIssue, Title: "x",
					URL:    "https://github.com/gnolang/gno/issues/100",
					Author: "moul", UpdatedAt: time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC),
					Comments: 5},
			}},
		},
	}
	var buf bytes.Buffer
	if err := RenderJSON(&buf, r); err != nil {
		t.Fatal(err)
	}
	var back map[string]any
	if err := json.Unmarshal(buf.Bytes(), &back); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if back["window_days"].(float64) != 30 {
		t.Errorf("window_days mismatch: %v", back["window_days"])
	}
	secs := back["sections"].([]any)
	if len(secs) != 1 {
		t.Fatalf("sections len: %d", len(secs))
	}
}

func TestRenderMarkdown(t *testing.T) {
	r := Report{
		GeneratedAt: time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
		WindowDays:  30,
		Sections: []Section{
			{Name: "Hot", Entries: []Entry{
				{Repo: "gnolang/gno", Number: 100, Kind: KindIssue, Title: "x",
					URL:    "https://github.com/gnolang/gno/issues/100",
					Author: "moul", UpdatedAt: time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC),
					Comments: 5},
			}},
			{Name: "Depends on other core", Entries: []Entry{
				{Repo: "gnolang/gno", Number: 101, Kind: KindPR, Title: "y",
					URL:    "https://github.com/gnolang/gno/pull/101",
					Author: "alice", UpdatedAt: time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC),
					RequestedReviewer: []string{"thehowl"}},
			}},
		},
	}
	old := now
	now = func() time.Time { return time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC) }
	defer func() { now = old }()

	want, err := os.ReadFile("testdata/golden_md.txt")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := RenderMarkdown(&buf, r); err != nil {
		t.Fatal(err)
	}
	if buf.String() != string(want) {
		t.Errorf("markdown mismatch:\n--- want ---\n%s--- got ---\n%s", want, buf.String())
	}
}
