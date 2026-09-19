package main

import (
	"testing"
	"time"
)

func fixedNow(t *testing.T, s string) {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	old := now
	now = func() time.Time { return parsed }
	t.Cleanup(func() { now = old })
}

func TestIsHot(t *testing.T) {
	fixedNow(t, "2026-05-20T00:00:00Z")

	cases := []struct {
		name string
		e    Entry
		want bool
	}{
		{
			"5 recent human comments",
			Entry{RecentComments: []Comment{
				{Author: "a", CreatedAt: mustTime("2026-05-18T00:00:00Z")},
				{Author: "b", CreatedAt: mustTime("2026-05-18T00:00:00Z")},
				{Author: "c", CreatedAt: mustTime("2026-05-18T00:00:00Z")},
				{Author: "d", CreatedAt: mustTime("2026-05-18T00:00:00Z")},
				{Author: "e", CreatedAt: mustTime("2026-05-18T00:00:00Z")},
			}, Comments: 5},
			true,
		},
		{
			"bot comments don't count",
			Entry{RecentComments: []Comment{
				{Author: "bot[bot]", IsBot: true, CreatedAt: mustTime("2026-05-18T00:00:00Z")},
				{Author: "bot[bot]", IsBot: true, CreatedAt: mustTime("2026-05-18T00:00:00Z")},
				{Author: "bot[bot]", IsBot: true, CreatedAt: mustTime("2026-05-18T00:00:00Z")},
				{Author: "bot[bot]", IsBot: true, CreatedAt: mustTime("2026-05-18T00:00:00Z")},
				{Author: "bot[bot]", IsBot: true, CreatedAt: mustTime("2026-05-18T00:00:00Z")},
			}},
			false,
		},
		{"3 reactions", Entry{Reactions: 3}, true},
		{"2 reactions, no comments", Entry{Reactions: 2}, false},
		{
			"comments older than 7d don't count",
			Entry{RecentComments: []Comment{
				{Author: "a", CreatedAt: mustTime("2026-05-01T00:00:00Z")},
				{Author: "b", CreatedAt: mustTime("2026-05-01T00:00:00Z")},
				{Author: "c", CreatedAt: mustTime("2026-05-01T00:00:00Z")},
				{Author: "d", CreatedAt: mustTime("2026-05-01T00:00:00Z")},
				{Author: "e", CreatedAt: mustTime("2026-05-01T00:00:00Z")},
			}, Comments: 5},
			false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isHot(c.e); got != c.want {
				t.Errorf("isHot=%v want %v", got, c.want)
			}
		})
	}
}

func TestIsStale(t *testing.T) {
	fixedNow(t, "2026-05-20T00:00:00Z")

	cases := []struct {
		name string
		e    Entry
		want bool
	}{
		{"updated 20d ago", Entry{UpdatedAt: mustTime("2026-04-30T00:00:00Z")}, true},
		{"updated 5d ago", Entry{UpdatedAt: mustTime("2026-05-15T00:00:00Z")}, false},
		{"exactly 14d ago", Entry{UpdatedAt: mustTime("2026-05-06T00:00:00Z")}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isStale(c.e); got != c.want {
				t.Errorf("isStale=%v want %v", got, c.want)
			}
		})
	}
}

func mustTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestDependsOn(t *testing.T) {
	fixedNow(t, "2026-05-20T00:00:00Z")

	cases := []struct {
		name   string
		e      Entry
		handle string
		want   bool
	}{
		{"assignee match", Entry{Assignees: []string{"jaekwon"}}, "jaekwon", true},
		{"requested reviewer match", Entry{RequestedReviewer: []string{"moul"}}, "moul", true},
		{"mention in last comment",
			Entry{RecentComments: []Comment{{Body: "could @jaekwon take a look?"}}},
			"jaekwon", true},
		{"unrelated", Entry{Assignees: []string{"alice"}}, "jaekwon", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := dependsOn(c.e, c.handle); got != c.want {
				t.Errorf("dependsOn(%q)=%v want %v", c.handle, got, c.want)
			}
		})
	}
}

func TestDependsOnOtherCore(t *testing.T) {
	fixedNow(t, "2026-05-20T00:00:00Z")
	e := Entry{RequestedReviewer: []string{"thehowl"}}
	who, ok := dependsOnOtherCore(e)
	if !ok || who != "thehowl" {
		t.Errorf("dependsOnOtherCore=%q,%v want thehowl,true", who, ok)
	}
	e2 := Entry{RequestedReviewer: []string{"alice"}}
	if _, ok := dependsOnOtherCore(e2); ok {
		t.Errorf("dependsOnOtherCore should be false for non-core")
	}
}

func TestIsReadyToMerge(t *testing.T) {
	cases := []struct {
		name string
		e    Entry
		want bool
	}{
		{
			"approved + green + mergeable",
			Entry{Kind: KindPR, Mergeable: "MERGEABLE", StatusCheckRollup: "SUCCESS",
				Reviews: []Review{{Author: "moul", State: "APPROVED"}}},
			true,
		},
		{
			"approved but draft",
			Entry{Kind: KindPR, IsDraft: true, Mergeable: "MERGEABLE", StatusCheckRollup: "SUCCESS",
				Reviews: []Review{{Author: "moul", State: "APPROVED"}}},
			false,
		},
		{
			"approved but unresolved changes-requested from another reviewer",
			Entry{Kind: KindPR, Mergeable: "MERGEABLE", StatusCheckRollup: "SUCCESS",
				Reviews: []Review{
					{Author: "moul", State: "APPROVED", SubmittedAt: mustTime("2026-05-19T00:00:00Z")},
					{Author: "alice", State: "CHANGES_REQUESTED", SubmittedAt: mustTime("2026-05-18T00:00:00Z")},
				}},
			false,
		},
		{
			"approved + reviewer later approved (overriding own changes-requested)",
			Entry{Kind: KindPR, Mergeable: "MERGEABLE", StatusCheckRollup: "SUCCESS",
				Reviews: []Review{
					{Author: "alice", State: "CHANGES_REQUESTED", SubmittedAt: mustTime("2026-05-18T00:00:00Z")},
					{Author: "alice", State: "APPROVED", SubmittedAt: mustTime("2026-05-19T00:00:00Z")},
				}},
			true,
		},
		{
			"mergeable unknown",
			Entry{Kind: KindPR, Mergeable: "UNKNOWN", StatusCheckRollup: "SUCCESS",
				Reviews: []Review{{Author: "moul", State: "APPROVED"}}},
			false,
		},
		{
			"CI failing",
			Entry{Kind: KindPR, Mergeable: "MERGEABLE", StatusCheckRollup: "FAILURE",
				Reviews: []Review{{Author: "moul", State: "APPROVED"}}},
			false,
		},
		{"issue, not a PR", Entry{Kind: KindIssue}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isReadyToMerge(c.e); got != c.want {
				t.Errorf("isReadyToMerge=%v want %v", got, c.want)
			}
		})
	}
}

func TestIsStuck(t *testing.T) {
	fixedNow(t, "2026-05-20T00:00:00Z")
	cases := []struct {
		name string
		e    Entry
		want bool
	}{
		{
			"opened 35d ago, review requested, no update 10d",
			Entry{
				CreatedAt:       mustTime("2026-04-15T00:00:00Z"),
				UpdatedAt:       mustTime("2026-05-10T00:00:00Z"),
				ReviewRequested: true,
			},
			true,
		},
		{
			"opened 35d ago but no review requested",
			Entry{
				CreatedAt:       mustTime("2026-04-15T00:00:00Z"),
				UpdatedAt:       mustTime("2026-05-10T00:00:00Z"),
				ReviewRequested: false,
			},
			false,
		},
		{
			"opened 35d ago, review requested, updated yesterday",
			Entry{
				CreatedAt:       mustTime("2026-04-15T00:00:00Z"),
				UpdatedAt:       mustTime("2026-05-19T00:00:00Z"),
				ReviewRequested: true,
			},
			false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isStuck(c.e); got != c.want {
				t.Errorf("isStuck=%v want %v", got, c.want)
			}
		})
	}
}

func TestIsNewContributor(t *testing.T) {
	cases := []struct {
		name string
		e    Entry
		want bool
	}{
		{"first timer", Entry{AuthorAssociation: "FIRST_TIMER"}, true},
		{"first time contributor", Entry{AuthorAssociation: "FIRST_TIME_CONTRIBUTOR"}, true},
		{"none association", Entry{AuthorAssociation: "NONE"}, true},
		{"contributor association, old account",
			Entry{AuthorAssociation: "CONTRIBUTOR", AuthorAccountAge: 365 * 24 * time.Hour}, false},
		{"contributor association, young account",
			Entry{AuthorAssociation: "CONTRIBUTOR", AuthorAccountAge: 30 * 24 * time.Hour}, true},
		{"bot author excluded",
			Entry{AuthorAssociation: "NONE", AuthorIsBot: true}, false},
		{"member", Entry{AuthorAssociation: "MEMBER"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isNewContributor(c.e); got != c.want {
				t.Errorf("isNewContributor=%v want %v", got, c.want)
			}
		})
	}
}

func TestClassify(t *testing.T) {
	fixedNow(t, "2026-05-20T00:00:00Z")

	entries := []Entry{
		// Excluded (label).
		{Number: 1, Labels: []string{"wontfix"}, UpdatedAt: mustTime("2026-04-01T00:00:00Z")},
		// Hot (5 recent comments).
		{Number: 2, UpdatedAt: mustTime("2026-05-18T00:00:00Z"),
			RecentComments: []Comment{
				{Author: "a", CreatedAt: mustTime("2026-05-18T00:00:00Z")},
				{Author: "b", CreatedAt: mustTime("2026-05-18T00:00:00Z")},
				{Author: "c", CreatedAt: mustTime("2026-05-18T00:00:00Z")},
				{Author: "d", CreatedAt: mustTime("2026-05-18T00:00:00Z")},
				{Author: "e", CreatedAt: mustTime("2026-05-18T00:00:00Z")},
			}},
		// Stale fall-through (no other category).
		{Number: 3, UpdatedAt: mustTime("2026-04-30T00:00:00Z")},
		// Stale + Depends on Jae (so does NOT appear in Stale).
		{Number: 4, UpdatedAt: mustTime("2026-04-30T00:00:00Z"),
			Assignees: []string{"jaekwon"}},
	}
	r := Classify(entries)
	got := map[string][]int{}
	for _, s := range r.Sections {
		for _, e := range s.Entries {
			got[s.Name] = append(got[s.Name], e.Number)
		}
	}
	if !equalInts(got["Hot"], []int{2}) {
		t.Errorf("Hot=%v want [2]", got["Hot"])
	}
	if !equalInts(got["Stale"], []int{3}) {
		t.Errorf("Stale=%v want [3]", got["Stale"])
	}
	if !equalInts(got["Depends on @jaekwon"], []int{4}) {
		t.Errorf("Depends on @jaekwon=%v want [4]", got["Depends on @jaekwon"])
	}
	if len(got["Hot"])+len(got["Stale"])+len(got["Depends on @jaekwon"]) != 3 {
		t.Errorf("entry #1 should be excluded; total sections entries: %v", got)
	}
}

func TestClassify_DraftNotStale(t *testing.T) {
	fixedNow(t, "2026-05-20T00:00:00Z")

	entries := []Entry{
		// Draft PR with old UpdatedAt — must NOT appear in Stale.
		{Number: 10, Kind: KindPR, IsDraft: true, UpdatedAt: mustTime("2026-04-01T00:00:00Z")},
	}
	r := Classify(entries)
	for _, s := range r.Sections {
		if s.Name == "Stale" {
			for _, e := range s.Entries {
				if e.Number == 10 {
					t.Errorf("draft PR #10 must not appear in Stale")
				}
			}
		}
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
