package mermaid

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimelineErrors(t *testing.T) {
	renderErrors(t, map[string][2]string{
		"no periods":     {"timeline\ntitle x\nsection y", "without periods"},
		"stray event":    {"timeline\n: orphan", "event before any period"},
		"option":         {"timeline LR\n2024 : x", "unsupported timeline option"},
		"colon period":   {"timeline\n10:00 : x", "colon in a period"},
		"trailing colon": {"timeline\n2020 : a\n2021 :", "colon in a period"},
		"empty event":    {"timeline\n2020 : : b", "empty event"},
		"colon section":  {"timeline\n2020 : a\nsection Phase 1 : launch", "colon in a section"},
	})
}

func TestTimelineParse(t *testing.T) {
	var m meta
	d, err := parseTimeline("", toLines(
		"2023 : early",
		"Section Later",
		"2024 : a : b",
		"     : c",
		"2025",
		"2026 : at 10:30 : https://gno.land : x :",
		"Title Gno",
		"# a comment",
	), &m)
	require.NoError(t, err)
	tl := d.(*timeline)
	require.Len(t, tl.periods, 4)
	assert.Equal(t, "Gno", m.title)
	assert.Equal(t, -1, tl.periods[0].section, "a period before any section has none")
	assert.Equal(t, 0, tl.periods[1].section)
	assert.Equal(t, [][]string{{"a"}, {"b"}, {"c"}}, tl.periods[1].events)
	assert.Empty(t, tl.periods[2].events)
	assert.Equal(t, [][]string{{"at 10:30"}, {"https://gno.land"}, {"x :"}}, tl.periods[3].events,
		"only a colon followed by a space separates events")
}

func TestTimelineComments(t *testing.T) {
	// Only a line starting with a # that starts no entity code is a comment;
	// elsewhere a # is text.
	d, err := parseTimeline("", toLines(
		"2021 : a",
		"#9829; love : x",
		": y",
		"#35;1 : launch",
		"#quot;Q1#quot; : launch",
		"# 2022 : dropped",
		"PR #6543 : merged",
	), new(meta))
	require.NoError(t, err)
	var got []string
	for _, p := range d.(*timeline).periods {
		got = append(got, fmt.Sprint(p.label, p.events))
	}
	assert.Equal(t, []string{
		"[2021] [[a]]",
		"[♥ love] [[x] [y]]",
		"[#1] [[launch]]",
		`["Q1"] [[launch]]`,
		"[PR #6543] [[merged]]",
	}, got)
}
