package mermaid

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKanbanErrors(t *testing.T) {
	renderErrors(t, map[string][2]string{
		"no columns":      {"kanban", "without columns"},
		"option":          {"kanban LR\ncol", "unsupported kanban option"},
		"unclosed label":  {"kanban\ncol[Todo\n  card", "malformed label"},
		"empty label":     {"kanban\ncol[]", "empty label"},
		"unclosed meta":   {"kanban\ncol\n  card@{ priority: 'High'", "unclosed metadata"},
		"malformed meta":  {"kanban\ncol\n  card@{ High }", "malformed metadata"},
		"column metadata": {"kanban\ncol@{ ticket: 1 }", "metadata on a column"},
		"shape":           {"kanban\ncol\n  t1(Write docs)", "shapes"},
		"double brackets": {"kanban\ncol\n  t1[[x]]", "malformed label"},
		"icon":            {"kanban\ncol\n  card\n  ::icon(fa fa-book)", "icons"},
		"class":           {"kanban\ncol\n  card:::urgent", "icons and classes"},
		"shallower line":  {"kanban\n  Todo\n    Write docs\n Release", "indented less than the columns"},
		"unclosed quote":  {"kanban\ncol\n  card@{ label: \"a, b }", "unclosed quote"},
	})
}

func TestKanbanMetadata(t *testing.T) {
	// Commas and colons inside quotes are part of the value; the label is
	// decoded once, so an encoded tag stays text.
	for meta, want := range map[string]kanbanCard{
		`label: "Standup, 10:00"`:          {label: []string{"Standup, 10:00"}, priority: -1},
		`label: "a, priority: High, b: c"`: {label: []string{"a, priority: High, b: c"}, priority: -1},
		`label: "one#60;br#62;two"`:        {label: []string{"one<br>two"}, priority: -1},
		`ticket: 'T·', assigned: '·bob'`:   {label: []string{"x"}, details: "T· · ·bob", priority: -1},
		`assigned: 'bob', priority: 'Low'`: {label: []string{"x"}, details: "bob", priority: 1},
	} {
		card, err := parseKanbanNode(line{text: "x@{ " + meta + " }"})
		require.NoError(t, err, meta)
		assert.Equal(t, want, card, meta)
	}
}

func TestKanbanColumnLevel(t *testing.T) {
	// The first line sets the column level, as in Mermaid: a deeper line is
	// a card of the last column, at any depth.
	d, err := parseKanban("", []line{
		{indent: 2, text: "Todo"},
		{indent: 4, text: "a"},
		{indent: 3, text: "b"},
		{indent: 2, text: "Done"},
	}, &meta{})
	require.NoError(t, err)
	k := d.(*kanban)
	require.Len(t, k.columns, 2)
	assert.Len(t, k.columns[0].cards, 2)
}

func TestKanbanParse(t *testing.T) {
	d, err := parseKanban("", []line{
		{indent: 2, text: "todo[To do]"},
		{indent: 4, text: "a[First]@{ ticket: T-1, assigned: 'ann', priority: 'Very High' }"},
		{indent: 4, text: "plain card"},
		{indent: 4, text: `q["Quoted"]@{ priority: 'urgent' }`},
		{indent: 2, text: "done"},
		{indent: 6, text: "b[Second]@{ assigned: \"bo\", other: x, label: 'Renamed' }"},
	}, &meta{})
	require.NoError(t, err)
	k := d.(*kanban)
	require.Len(t, k.columns, 2)
	assert.Equal(t, []string{"To do"}, k.columns[0].label)
	assert.Equal(t, kanbanCard{label: []string{"First"}, details: "T-1 · ann", priority: 4}, k.columns[0].cards[0])
	assert.Equal(t, kanbanCard{label: []string{"plain card"}, priority: -1}, k.columns[0].cards[1])
	assert.Equal(t, kanbanCard{label: []string{"Quoted"}, priority: -1}, k.columns[0].cards[2], "an unknown priority draws no stripe")
	assert.Equal(t, []string{"done"}, k.columns[1].label)
	assert.Equal(t, kanbanCard{label: []string{"Renamed"}, details: "bo", priority: -1}, k.columns[1].cards[0])
}

func TestKanbanTabs(t *testing.T) {
	// A tab is one level, as in Mermaid: a card under a tab-indented column
	// needs more than one character of indentation.
	out, _, err := Render("kanban\n\tTodo\n\t  card\n")
	require.NoError(t, err)
	assert.Contains(t, string(out), ">card<")
	assert.Equal(t, 1, strings.Count(string(out), "gno-mermaid__box"), "one lane")
}
