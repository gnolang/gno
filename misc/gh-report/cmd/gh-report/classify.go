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
