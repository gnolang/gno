package mermaid

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimelineErrors(t *testing.T) {
	renderErrors(t, map[string][2]string{
		"no periods":   {"timeline\ntitle x\nsection y", "without periods"},
		"stray event":  {"timeline\n: orphan", "event before any period"},
		"option":       {"timeline LR\n2024 : x", "unsupported timeline option"},
		"colon period": {"timeline\n10:00 : x", "colon in a period"},
	})
}

func TestTimelineParse(t *testing.T) {
	var m meta
	d, err := parseTimeline("", []line{
		{text: "2023 : early"},
		{text: "Section Later"},
		{text: "2024 : a : b"},
		{text: "     : c"},
		{text: "2025"},
		{text: "2026 : at 10:30 : https://gno.land"},
		{text: "Title Gno"},
	}, &m)
	require.NoError(t, err)
	tl := d.(*timeline)
	require.Len(t, tl.periods, 4)
	assert.Equal(t, "Gno", m.title)
	assert.Equal(t, -1, tl.periods[0].section, "a period before any section has none")
	assert.Equal(t, 0, tl.periods[1].section)
	assert.Equal(t, [][]string{{"a"}, {"b"}, {"c"}}, tl.periods[1].events)
	assert.Empty(t, tl.periods[2].events)
	assert.Equal(t, [][]string{{"at 10:30"}, {"https://gno.land"}}, tl.periods[3].events,
		"only a colon followed by a space separates events")
}
