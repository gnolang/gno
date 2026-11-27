package mermaid

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPieErrors(t *testing.T) {
	renderErrors(t, map[string][2]string{
		"zero total":   {"pie\n\"a\" : 0", "positive"},
		"no slices":    {"pie title x", "positive"},
		"negative":     {"pie\n\"a\" : -1", "non-negative"},
		"not a number": {"pie\n\"a\" : many", "non-negative"},
		"underscores":  {"pie\n\"a\" : 1_000", "non-negative"},
		"unquoted":     {"pie\na : 1", "unsupported statement"},
		"unclosed":     {"pie\n\"a : 1", "unclosed quote"},
		"no value":     {"pie\n\"a\"", "without a value"},
		"option":       {"pie explode", "unsupported pie option"},
	})
}

func TestPieWhole(t *testing.T) {
	// A single slice is a full circle: an arc cannot start and end at the
	// same point. A zero slice draws nothing but keeps its legend row.
	out, _, err := Render("pie\n\"all\" : 3\n\"none\" : 0")
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, `<circle class="gno-mermaid__slice gno-mermaid__c0"`)
	assert.NotContains(t, s, "<path")
	assert.Equal(t, 1, strings.Count(s, "gno-mermaid__slice"))
	assert.Contains(t, s, "all 100%")
	assert.Contains(t, s, "none 0%")
}

func TestPieNearlyWhole(t *testing.T) {
	// A slice a hair short of a full turn would start and end at the same
	// rounded point, which SVG does not draw: it goes through its middle.
	out, _, err := Render("pie\n\"big\" : 999999\n\"small\" : 1")
	require.NoError(t, err)
	big := regexp.MustCompile(`d="M126 126L126 16A110 110 0 0 1 ([\d.]+ [\d.]+)A110 110 0 0 1 126 16Z"`)
	m := big.FindSubmatch(out)
	require.NotNil(t, m, "%s", out)
	assert.Equal(t, "126 236", string(m[1]), "the middle is the bottom of the circle")
}

func TestPieDuplicates(t *testing.T) {
	// Mermaid keeps the first slice of a label.
	d, err := parsePie("", toLines(`"a" : 1`, `"b" : 1`, `"a" : 8`), &meta{})
	require.NoError(t, err)
	p := d.(*pie)
	assert.Equal(t, []slice{{"a", 1}, {"b", 1}}, p.slices)
	assert.Equal(t, 2.0, p.total)
}

func TestPieColors(t *testing.T) {
	// The last slice never takes the color of the first, which it touches.
	p := &pie{slices: make([]slice, len(palette)+1)}
	assert.Equal(t, color(0), p.sliceColor(0))
	assert.NotEqual(t, p.sliceColor(0), p.sliceColor(len(palette)))
	assert.NotEqual(t, p.sliceColor(len(palette)-1), p.sliceColor(len(palette)))
}

func TestPieLegend(t *testing.T) {
	d, err := parsePie("showData", toLines(`"a" : -0`, `"b" : 1e-300`, `"c" : 1`, `"d" : 0.0000001`), &meta{})
	require.NoError(t, err)
	p := d.(*pie)
	assert.Equal(t, "a 0% (0)", p.legend(p.slices[0]))
	assert.Equal(t, "b 0% (1e-300)", p.legend(p.slices[1]))
	assert.Equal(t, "d 0% (1e-7)", p.legend(p.slices[3]), "as JavaScript writes it")
}

func TestPercent(t *testing.T) {
	assert.Equal(t, "65.6%", percent(42, 64))
	assert.Equal(t, "50%", percent(1, 2))
	assert.Equal(t, "33.3%", percent(1, 3))
}
