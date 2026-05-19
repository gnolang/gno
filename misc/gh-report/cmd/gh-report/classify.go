package main

import (
	"strings"
	"time"
)

// Tune these in code as we use the tool.
const (
	WindowDays           = 30
	StaleDays            = 14
	StuckOpenDays        = 30
	StuckNoUpdateDays    = 7
	HotRecentDays        = 7
	HotComments          = 5
	HotReactions         = 3
	NewContribAccountDays = 90
)

var (
	WatchJaekwon = "jaekwon"
	WatchMoul    = "moul"
	OtherCore    = []string{"zivkovicmilos", "thehowl", "leohhhn"}

	ExcludeLabels = []string{"wontfix", "duplicate", "invalid"}
)

// now is overridable in tests for deterministic age calculations.
var now = func() time.Time { return time.Now() }

func ageDays(t time.Time) int {
	return int(now().Sub(t).Hours() / 24)
}

func hasAny(list []string, target string) bool {
	for _, s := range list {
		if strings.EqualFold(s, target) {
			return true
		}
	}
	return false
}

func hasAnyOf(list, targets []string) bool {
	for _, t := range targets {
		if hasAny(list, t) {
			return true
		}
	}
	return false
}

func excluded(e Entry) bool {
	for _, l := range e.Labels {
		if hasAny(ExcludeLabels, l) {
			return true
		}
	}
	return false
}

func recentHumanComments(e Entry) int {
	cutoff := now().AddDate(0, 0, -HotRecentDays)
	n := 0
	for _, c := range e.RecentComments {
		if c.IsBot {
			continue
		}
		if c.CreatedAt.After(cutoff) {
			n++
		}
	}
	return n
}

func isHot(e Entry) bool {
	if recentHumanComments(e) >= HotComments {
		return true
	}
	if e.Reactions >= HotReactions {
		return true
	}
	return false
}

func isStale(e Entry) bool {
	return ageDays(e.UpdatedAt) >= StaleDays
}

// dependsOn returns true if entry e is gated by user `handle`:
// assignee match, requested reviewer match, or @handle mention in last comments.
func dependsOn(e Entry, handle string) bool {
	if hasAny(e.Assignees, handle) {
		return true
	}
	if hasAny(e.RequestedReviewer, handle) {
		return true
	}
	needle := "@" + strings.ToLower(handle)
	for _, c := range e.RecentComments {
		if strings.Contains(strings.ToLower(c.Body), needle) {
			return true
		}
	}
	return false
}

// dependsOnOtherCore returns the matched core handle (other than jaekwon/moul) if any.
func dependsOnOtherCore(e Entry) (string, bool) {
	for _, h := range OtherCore {
		if dependsOn(e, h) {
			return h, true
		}
	}
	return "", false
}

func isReadyToMerge(e Entry) bool {
	if e.Kind != KindPR || e.IsDraft {
		return false
	}
	if e.Mergeable != "MERGEABLE" {
		return false
	}
	if e.StatusCheckRoll != "" && e.StatusCheckRoll != "SUCCESS" {
		return false
	}
	// Latest review per author. If any latest is CHANGES_REQUESTED, block.
	latest := map[string]Review{}
	for _, r := range e.Reviews {
		if r.State == "" {
			continue
		}
		if cur, ok := latest[r.Author]; !ok || r.SubmittedAt.After(cur.SubmittedAt) {
			latest[r.Author] = r
		}
	}
	approved := false
	for _, r := range latest {
		if r.State == "CHANGES_REQUESTED" {
			return false
		}
		if r.State == "APPROVED" {
			approved = true
		}
	}
	return approved
}

func isStuck(e Entry) bool {
	if !e.ReviewRequested {
		return false
	}
	if ageDays(e.CreatedAt) <= StuckOpenDays {
		return false
	}
	if ageDays(e.UpdatedAt) < StuckNoUpdateDays {
		return false
	}
	return true
}
