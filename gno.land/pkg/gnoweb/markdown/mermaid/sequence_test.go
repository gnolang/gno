package mermaid

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSequenceErrors(t *testing.T) {
	many := "sequenceDiagram\n"
	for i := 0; i <= maxParticipants; i++ {
		many += "participant P" + strings.Repeat("x", i) + "\n"
	}
	renderErrors(t, map[string][2]string{
		"no participants":   {"sequenceDiagram\ntitle x", "no participants"},
		"unclosed block":    {"sequenceDiagram\nloop forever\nA->>B: x", "loop block is not closed"},
		"stray end":         {"sequenceDiagram\nA->>B: x\nend", "unexpected end"},
		"else outside":      {"sequenceDiagram\nloop x\nelse y\nend", `"else" outside its block`},
		"box":               {"sequenceDiagram\nbox Aqua Group\nparticipant A\nend", `unsupported statement "box"`},
		"rect":              {"sequenceDiagram\nrect rgb(0, 0, 0)\nA->>B: x\nend", `unsupported statement "rect"`},
		"link":              {"sequenceDiagram\nA->>B: x\nlink A: Docs @ https://example.com", `unsupported statement "link"`},
		"create":            {"sequenceDiagram\nA->>B: x\ncreate participant C", `unsupported statement "create"`},
		"no sender":         {"sequenceDiagram\n->>B: x", "participant name"},
		"no colon":          {"sequenceDiagram\nA->>B hello", "unsupported statement"},
		"note position":     {"sequenceDiagram\nNote under A: x", "note position"},
		"note empty name":   {"sequenceDiagram\nNote over A,: x", "participant name"},
		"note three":        {"sequenceDiagram\nNote over A,B,C: x", "too many"},
		"note left two":     {"sequenceDiagram\nNote left of A,B: x", "too many"},
		"typed":             {"sequenceDiagram\nparticipant A@{ \"type\": \"database\" }", "participant name"},
		"arrow in name":     {"sequenceDiagram\nweb-xhr->>B: q", "participant name"},
		"autonumber":        {"sequenceDiagram\nautonumber one\nA->>B: x", "autonumber"},
		"autonumber big":    {"sequenceDiagram\nautonumber 99999999999\nA->>B: x", "autonumber"},
		"activate unknown":  {"sequenceDiagram\nA->>B: x\nactivate C", "unknown participant"},
		"non-ASCII keyword": {"sequenceDiagram\nNote r\u0130ght of A: x", "note"},
		"header":            {"sequenceDiagram A->>B: x", "sequenceDiagram line"},
		"participants":      {many, "participants"},
		"unclosed accDescr": {"sequenceDiagram\naccDescr {\n  text", "accDescr block"},
	})
}

func parseSeq(t *testing.T, lines ...string) (*sequence, meta) {
	t.Helper()
	var m meta
	d, err := parseSequence("", toLines(lines...), &m)
	require.NoError(t, err)
	return d.(*sequence), m
}

func TestSequenceArrows(t *testing.T) {
	// Each arrow token parses to its own style, a longer one never read as a
	// shorter one it starts with.
	for tok, want := range map[string]seqArrow{
		"->":     {},
		"-->":    {dashed: true},
		"->>":    {head: '>'},
		"-->>":   {dashed: true, head: '>'},
		"-x":     {head: 'x'},
		"--x":    {dashed: true, head: 'x'},
		"-)":     {head: ')'},
		"--)":    {dashed: true, head: ')'},
		"<<->>":  {head: '>', both: true},
		"<<-->>": {dashed: true, head: '>', both: true},
	} {
		s, _ := parseSeq(t, "Alice"+tok+"+Bob: hi")
		require.Len(t, s.parts, 2, tok)
		assert.Equal(t, want, s.events[0].arrow, tok)
	}
}

func TestSequenceNames(t *testing.T) {
	s, _ := parseSeq(t, "gov-dao->>tre-asury: x", "participant C AS Carol", "Alice Smith->>C: y", "participant ȺȺȺȺ as x")
	var labels [][]string
	for _, p := range s.parts {
		labels = append(labels, p.label)
	}
	assert.Equal(t, [][]string{{"gov-dao"}, {"tre-asury"}, {"Carol"}, {"Alice Smith"}, {"x"}}, labels, "an alias after a name that lowercases longer")
}

func TestSequenceStatements(t *testing.T) {
	// A semicolon separates statements, except the one ending an entity code.
	s, _ := parseSeq(t, "participant A;", "A->>B: hi; B->>A: a#59;b")
	require.Len(t, s.parts, 2)
	require.Len(t, s.events, 2)
	assert.Equal(t, []string{"hi"}, s.events[0].text)
	assert.Equal(t, []string{"a;b"}, s.events[1].text)
}

func TestSequenceAutonumber(t *testing.T) {
	// Mermaid counts every message, numbered or not; a bare autonumber
	// resumes the count with the current step.
	s, m := parseSeq(t,
		"A->>B: 1",
		"autonumber",
		"A->>B: 2",
		"autonumber 10 5",
		"A->>B: 10",
		"autonumber off",
		"A->>B: 15",
		"autonumber",
		"A->>B: 20",
		"Title Numbers",
	)
	var nums []string
	for _, e := range s.events {
		nums = append(nums, e.num)
	}
	assert.Equal(t, []string{"", "2", "10", "", "20"}, nums)
	assert.Equal(t, "Numbers", m.title, "keywords ignore case")

	// As in Mermaid, 0 keeps the current start or step, and a start alone
	// resets the step to 1.
	for src, want := range map[string][]string{
		"autonumber 0\nA->>B: x\nA->>B: y":                             {"1", "2"},
		"autonumber 10 0\nA->>B: x\nA->>B: y":                          {"10", "11"},
		"autonumber 1 10\nA->>B: x\nautonumber 50\nA->>B: y\nA->>B: z": {"1", "50", "51"},
	} {
		lines := strings.Split(src, "\n")
		s, _ := parseSeq(t, lines...)
		var nums []string
		for _, e := range s.events {
			nums = append(nums, e.num)
		}
		assert.Equal(t, want, nums, src)
	}
}

func TestSequenceRedeclare(t *testing.T) {
	// As in Mermaid, declaring a participant again without an alias keeps
	// its shape; with one, it takes the new shape and label.
	s, _ := parseSeq(t, "A->>B: x", "actor A", "actor B as Bee")
	assert.False(t, s.parts[0].actor)
	assert.True(t, s.parts[1].actor)
	assert.Equal(t, []string{"Bee"}, s.parts[1].label)
}

func TestSequenceSelfBadge(t *testing.T) {
	// A two-headed self message draws its badge past the head at the top of
	// the loop, not over it.
	s, _ := parseSeq(t, "autonumber", "A<<->>A: ping")
	e := s.events[0]
	a := 0.0
	bx, bw := badgeX(e, a, a), badgeWidth(e.num)
	assert.GreaterOrEqual(t, bx-bw/2, a+10, "the head ends 10 px past the lifeline")
}
