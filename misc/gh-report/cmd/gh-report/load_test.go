package main

import (
	"os"
	"testing"
)

func TestLoadRepoFile(t *testing.T) {
	data, err := os.ReadFile("testdata/sample.json")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := LoadRepoJSON("gnolang/gno", data)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(entries))
	}

	issue := entries[0]
	if issue.Kind != KindIssue || issue.Number != 100 {
		t.Errorf("issue mismatch: %+v", issue)
	}
	if issue.Reactions != 4 || issue.Comments != 6 {
		t.Errorf("counts: reactions=%d comments=%d", issue.Reactions, issue.Comments)
	}
	if len(issue.Assignees) != 1 || issue.Assignees[0] != "moul" {
		t.Errorf("assignees: %v", issue.Assignees)
	}
	if len(issue.RecentComments) != 1 || issue.RecentComments[0].Body == "" {
		t.Errorf("recent comments missing")
	}

	pr := entries[1]
	if pr.Kind != KindPR || pr.Number != 200 {
		t.Errorf("PR mismatch: %+v", pr)
	}
	if pr.StatusCheckRoll != "SUCCESS" {
		t.Errorf("status: %q", pr.StatusCheckRoll)
	}
	if len(pr.Reviews) != 1 || pr.Reviews[0].State != "APPROVED" {
		t.Errorf("reviews: %v", pr.Reviews)
	}
	if !pr.ReviewRequested {
		t.Errorf("ReviewRequested should be true")
	}
	if pr.AuthorAssociation != "FIRST_TIME_CONTRIBUTOR" {
		t.Errorf("author assoc: %s", pr.AuthorAssociation)
	}
}
