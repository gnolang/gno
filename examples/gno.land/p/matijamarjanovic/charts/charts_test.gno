package charts

import (
	"strings"
	"testing"

	"gno.land/p/nt/uassert"
)

func TestGenerateLineChartMarkdown_Basic(t *testing.T) {
	values := []float64{1, 2, 3, 2, 1}
	result := GenerateLineChartMarkdown(values, 5, 6, "Test Line", "X", "Y")
	uassert.True(t, strings.Contains(result, "Test Line"))
	uassert.True(t, strings.Contains(result, "X"))
	uassert.True(t, strings.Contains(result, "Y"))
	uassert.True(t, strings.Contains(result, "*"))
}

func TestGenerateLineChartMarkdown_EmptyData(t *testing.T) {
	result := GenerateLineChartMarkdown([]float64{}, 5, 6, "Title", "X", "Y")
	uassert.Equal(t, "no data to display", result)
}

func TestGenerateLineChartMarkdown_BadSpacing(t *testing.T) {
	values := []float64{1, 2, 3}
	result := GenerateLineChartMarkdown(values, 5, 2, "Title", "X", "Y")
	uassert.Equal(t, "pointSpacing must be at least 6 to fit value labels", result)
}

func TestGenerateLineChartMarkdown_Normalization(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := GenerateLineChartMarkdown(values, 5, 25, "Norm", "X", "Y")
	println(result)
	uassert.True(t, strings.Contains(result, "Norm"))
	uassert.True(t, len(result) > 0)
}

func TestGenerateBarChartMarkdown_Basic(t *testing.T) {
	values := []float64{3, 5, 2}
	labels := []string{"A", "B", "C"}
	result := GenerateBarChartMarkdown(values, labels, 10, "Bar Test")
	uassert.True(t, strings.Contains(result, "Bar Test"))
	uassert.True(t, strings.Contains(result, "A"))
	uassert.True(t, strings.Contains(result, "B"))
	uassert.True(t, strings.Contains(result, "C"))
	uassert.True(t, strings.Contains(result, "█"))
}

func TestGenerateBarChartMarkdown_InvalidData(t *testing.T) {
	result := GenerateBarChartMarkdown([]float64{}, []string{}, 10, "Title")
	uassert.Equal(t, "invalid data for display", result)
	result2 := GenerateBarChartMarkdown([]float64{1, 2}, []string{"A"}, 10, "Title")
	uassert.Equal(t, "invalid data for display", result2)
}

func TestGenerateBarChartMarkdown_Widths(t *testing.T) {
	values := []float64{1, 2}
	labels := []string{"A", "B"}
	result := GenerateBarChartMarkdown(values, labels, 5, "Widths")
	uassert.True(t, strings.Contains(result, "Widths"))
}

func TestGenerateColumnChartMarkdown_Basic(t *testing.T) {
	values := []float64{2, 4, 3}
	result := GenerateColumnChartMarkdown(values, 3, "Col Test", "X", "Y")
	uassert.True(t, strings.Contains(result, "Col Test"))
	uassert.True(t, strings.Contains(result, "X"))
	uassert.True(t, strings.Contains(result, "Y"))
	uassert.True(t, strings.Contains(result, "|||||"))
}

func TestGenerateColumnChartMarkdown_EmptyData(t *testing.T) {
	result := GenerateColumnChartMarkdown([]float64{}, 3, "Title", "X", "Y")
	uassert.Equal(t, "no data to display", result)
}

func TestGenerateColumnChartMarkdown_Normalization(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6}
	result := GenerateColumnChartMarkdown(values, 3, "Norm", "X", "Y")
	uassert.True(t, strings.Contains(result, "Norm"))
}

func TestGenerateBarChart_SVG(t *testing.T) {
	values := []float64{3, 5, 2}
	labels := []string{"A", "B", "C"}
	result := GenerateBarChart(values, labels, 400, 240)
	uassert.True(t, strings.Contains(result, "data:image/svg+xml;base64"))
}

func TestGenerateLineChart_SVG(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	result := GenerateLineChart(values, 5, 480, 240, "X", "Y")
	uassert.True(t, strings.Contains(result, "data:image/svg+xml;base64"))
}

func TestGenerateColumnChart_SVG(t *testing.T) {
	values := []float64{2, 4, 3, 5}
	result := GenerateColumnChart(values, 4, 480, 240, "Idx", "Val")
	uassert.True(t, strings.Contains(result, "data:image/svg+xml;base64"))
}
