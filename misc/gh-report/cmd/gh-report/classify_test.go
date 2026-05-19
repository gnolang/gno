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
