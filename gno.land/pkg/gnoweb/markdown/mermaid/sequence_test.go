package mermaid

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSequenceErrors(t *testing.T) {
	var many strings.Builder
	many.WriteString("sequenceDiagram\n")
	for i := 0; i <= maxParticipants; i++ {
		many.WriteString("participant P" + strings.Repeat("x", i) + "\n")
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
		"participants":      {many.String(), "participants"},
		"unclosed accDescr": {"sequenceDiagram\naccDescr {\n  text", "accDescr block"},
		"# before ;":        {"sequenceDiagram\nA->>B: x # c; C->>D: y", `"#" before ";"`},
		"comment before ;":  {"sequenceDiagram\nA->>B: x; # c; C->>D: y", `"#" before ";"`},
		"entity at start":   {"sequenceDiagram\nA->>B: x\n#59;", "unsupported statement"},
		"entity name":       {"sequenceDiagram\nA->>B: x\n#quot;C#quot;->>B: y", "participant name"},
		"central to":        {"sequenceDiagram\nAlice->>()John: Hello", "central connection"},
		"central from":      {"sequenceDiagram\nAlice()->>John: x", "central connection"},
		"central both":      {"sequenceDiagram\nJohn()->>()Alice: x", "central connection"},
		"() in receiver":    {"sequenceDiagram\nMain->>init(): x", "central connection"},
		"() in sender":      {"sequenceDiagram\na()b->>C: x", "central connection"},
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

func TestSequenceComments(t *testing.T) {
	// A line or a statement starting with a # that does not start an entity
	// code is a comment, as in Mermaid; elsewhere a # is text.
	s, _ := parseSeq(t, "A->>B: hi", "# A->>C: old; C->>A: older", "C->>B: vote #42", "B->>A: x; # a comment")
	require.Len(t, s.parts, 3)
	require.Len(t, s.events, 3)
	assert.Equal(t, []string{"hi"}, s.events[0].text)
	assert.Equal(t, []string{"vote #42"}, s.events[1].text)
	assert.Equal(t, []string{"x"}, s.events[2].text)
}

func TestSequenceWrapPrefix(t *testing.T) {
	// Mermaid reads a wrap: or nowrap: prefix of a label as a setting, not
	// as text (its extractWrap).
	s, _ := parseSeq(t,
		"participant A as nowrap: Alice",
		"A->>John: wrap: Hello John",
		"Note right of John: nowrap:A note",
		"loop wrap: Every minute",
		"A->>John: :wrap:x",
		"end",
		"alt nowrap: yes",
		"A->>John: Wrap: unwrap: kept",
		"else wrap: no",
		"end",
	)
	assert.Equal(t, []string{"Alice"}, s.parts[0].label)
	var texts [][]string
	for _, e := range s.events {
		texts = append(texts, e.text)
	}
	assert.Equal(t, [][]string{
		{"Hello John"}, {"A note"}, {"Every minute"}, {"x"}, nil,
		{"yes"}, {"Wrap: unwrap: kept"}, {"no"}, nil,
	}, texts)
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

func TestSequenceSelfLabel(t *testing.T) {
	// A self message's label starts right of its autonumber badge, which a
	// two-headed loop or a long number puts under it, and the next
	// participant still clears the label.
	const text = "a label long enough to set the spacing"
	badge := regexp.MustCompile(`<rect class="gno-mermaid__head" x="([-\d.]+)" y="[-\d.]+" width="([\d.]+)"`)
	label := regexp.MustCompile(`<text x="([-\d.]+)" y="[-\d.]+">` + text + `<`)
	num := func(s string) float64 {
		v, err := strconv.ParseFloat(s, 64)
		require.NoError(t, err)
		return v
	}
	for _, src := range []string{"autonumber\nA<<->>A: ", "autonumber 1000000 1000000\nA->>A: "} {
		s, _ := parseSeq(t, strings.Split(src+text+"\nA->>B: x", "\n")...)
		var w svgWriter
		s.draw(&w)
		b, l := badge.FindStringSubmatch(w.b.String()), label.FindStringSubmatch(w.b.String())
		require.NotNil(t, b, src)
		require.NotNil(t, l, src)
		lx := num(l[1])
		assert.LessOrEqual(t, num(b[1])+num(b[2]), lx, "%s: the badge ends before the label", src)
		next := s.parts[1]
		assert.LessOrEqual(t, lx+textWidth(text, fontSize, false)+messageGap, next.x-next.w/2+0.01, "%s: B clears the label", src)
	}
}
