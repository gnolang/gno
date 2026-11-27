package mermaid

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXYChartErrors(t *testing.T) {
	renderErrors(t, map[string][2]string{
		"horizontal":     {"xychart-beta horizontal\nbar [1]", "unsupported xychart option"},
		"no data":        {"xychart-beta\ntitle x", "without data"},
		"numeric x-axis": {"xychart-beta\nx-axis 0 --> 10\nbar [1]", "numeric x-axis"},
		"bad range":      {"xychart-beta\ny-axis 10 --> 1\nbar [1]", "y-axis range"},
		"bad number":     {"xychart-beta\nbar [1, two]", "not a number"},
		"hex number":     {"xychart-beta\nbar [0x10]", "not a number"},
		"unclosed list":  {"xychart-beta\nbar [1, 2", "malformed list"},
		"unclosed quote": {"xychart-beta\nx-axis [\"a, b]\nbar [1]", "unclosed quote"},
		"named series":   {"xychart-beta\nbar \"a\" [1]", "unsupported bar statement"},
		"tiny range":     {"xychart-beta\ny-axis 0 --> 5e-324\nbar [0]", "y range too small"},
		"small range":    {"xychart-beta\ny-axis 0 --> 2e-15\nbar [0]", "y range too small"},
		"empty series":   {"xychart-beta\nbar []", "empty bar series"},
		"statement":      {"xychart-beta\nscatter [1]", `unsupported statement "scatter"`},
		"too wide": {"xychart-beta\nx-axis [" + strings.Repeat(strings.Repeat("w", 100)+",", 150) + "w]\nbar [1]",
			"larger than"},
	})
}

func parseXY(t *testing.T, lines ...string) (*xychart, meta) {
	t.Helper()
	var m meta
	d, err := parseXYChart("", toLines(lines...), &m)
	require.NoError(t, err)
	return d.(*xychart), m
}

func TestParseList(t *testing.T) {
	items, err := parseList(line{}, `[a, "b, c", d ,"e"]`)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b, c", "d", "e"}, items)
	items, err = parseList(line{}, `[]`)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestXYChartAxes(t *testing.T) {
	c, m := parseXY(t,
		`title: "Gas"`,
		`x-axis "Time [s]" [a, "b, c"]`,
		`y-axis Gas 0 --> 50`,
		`bar [1, 2, 3, 4]`,
	)
	assert.Equal(t, "Gas", m.title)
	assert.Equal(t, "Time [s]", c.xTitle, "a bracket in a quoted title is not the list")
	assert.Equal(t, []string{"a", "b, c"}, c.categories)
	assert.Equal(t, "Gas", c.yTitle)
	assert.Equal(t, []float64{1, 2}, c.series[0].values, "points past the categories are dropped")
	assert.Equal(t, 50.0, c.yMax)
}

func TestSplitTitle(t *testing.T) {
	for in, want := range map[string][2]string{
		`Month names [a, b]`:       {"Month names", "[a, b]"},
		`Revenue in USD 0 --> 100`: {"Revenue in USD", "0 --> 100"},
		`Temp -5-->5`:              {"Temp", "-5-->5"},
		`-5 --> 5`:                 {"", "-5 --> 5"},
		`[a]`:                      {"", "[a]"},
		`Just a title`:             {"Just a title", ""},
		`"Quoted [x]" [a]`:         {"Quoted [x]", "[a]"},
		`Amount #59; paid 1 --> 2`: {"Amount ; paid", "1 --> 2"},
	} {
		title, rest := splitTitle(in)
		assert.Equal(t, want, [2]string{title, rest}, in)
	}
}

func TestXYChartWideXTitle(t *testing.T) {
	// The x-axis title, centered under the plot, fits the drawing.
	const title = "A very long x axis title that is much wider than the plot"
	out, width, err := Render("xychart-beta\nx-axis \"" + title + "\" [a]\nbar [1]")
	require.NoError(t, err)
	m := regexp.MustCompile(`x="([\d.]+)" y="[\d.]+" text-anchor="middle" font-size="12">` + title).FindSubmatch(out)
	require.NotNil(t, m, "%s", out)
	x, _ := strconv.ParseFloat(string(m[1]), 64)
	half := textWidth(title, smallSize, true) / 2
	assert.GreaterOrEqual(t, x-half, pad-0.01)
	assert.LessOrEqual(t, x+half, width-pad+0.01)
}

func TestXYChartNegativeTicks(t *testing.T) {
	// The bottom of a negative range keeps its tick despite rounding.
	c, _ := parseXY(t, "y-axis -0.6 --> 0.4", "bar [0]")
	_, labels := c.ticks()
	assert.Equal(t, "-0.6", labels[0])
}

func TestXYChartAutoRange(t *testing.T) {
	c, _ := parseXY(t, "bar [-3, 7.5]")
	assert.Equal(t, -5.0, c.yMin)
	assert.Equal(t, 10.0, c.yMax)
	assert.Equal(t, []string{"1", "2"}, c.categories, "missing categories are numbered")
	_, labels := c.ticks()
	assert.Equal(t, []string{"-5", "0", "5", "10"}, labels)
}
