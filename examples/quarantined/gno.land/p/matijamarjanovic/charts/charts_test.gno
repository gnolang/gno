package charts

import (
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestGenerateLineChart_Basic(t *testing.T) {
	values := []float64{1, 2, 3, 2, 1}
	result := GenerateLineChart(values, 5, 6, "Test Line", "X", "Y")
	uassert.True(t, strings.Contains(result, "Test Line"))
	uassert.True(t, strings.Contains(result, "X"))
	uassert.True(t, strings.Contains(result, "Y"))
	uassert.True(t, strings.Contains(result, "*"))
}

func TestGenerateLineChart_EmptyData(t *testing.T) {
	result := GenerateLineChart([]float64{}, 5, 6, "Title", "X", "Y")
	uassert.Equal(t, "no data to display", result)
}

func TestGenerateLineChart_BadSpacing(t *testing.T) {
	values := []float64{1, 2, 3}
	result := GenerateLineChart(values, 5, 2, "Title", "X", "Y")
	uassert.Equal(t, "pointSpacing must be at least 6 to fit value labels", result)
}

func TestGenerateLineChart_Normalization(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := GenerateLineChart(values, 5, 25, "Norm", "X", "Y")
	println(result)
	uassert.True(t, strings.Contains(result, "Norm"))
	uassert.True(t, len(result) > 0)
}

func TestGenerateBarChart_Basic(t *testing.T) {
	values := []float64{3, 5, 2}
	labels := []string{"A", "B", "C"}
	result := GenerateBarChart(values, labels, 10, "Bar Test")
	uassert.True(t, strings.Contains(result, "Bar Test"))
	uassert.True(t, strings.Contains(result, "A"))
	uassert.True(t, strings.Contains(result, "B"))
	uassert.True(t, strings.Contains(result, "C"))
	uassert.True(t, strings.Contains(result, "█"))
}

func TestGenerateBarChart_InvalidData(t *testing.T) {
	result := GenerateBarChart([]float64{}, []string{}, 10, "Title")
	uassert.Equal(t, "invalid data for display", result)
	result2 := GenerateBarChart([]float64{1, 2}, []string{"A"}, 10, "Title")
	uassert.Equal(t, "invalid data for display", result2)
}

func TestGenerateBarChart_Widths(t *testing.T) {
	values := []float64{1, 2}
	labels := []string{"A", "B"}
	result := GenerateBarChart(values, labels, 5, "Widths")
	uassert.True(t, strings.Contains(result, "Widths"))
}

func TestGenerateColumnChart_Basic(t *testing.T) {
	values := []float64{2, 4, 3}
	result := GenerateColumnChart(values, 3, "Col Test", "X", "Y")
	uassert.True(t, strings.Contains(result, "Col Test"))
	uassert.True(t, strings.Contains(result, "X"))
	uassert.True(t, strings.Contains(result, "Y"))
	uassert.True(t, strings.Contains(result, "|||||"))
}

func TestGenerateColumnChart_EmptyData(t *testing.T) {
	result := GenerateColumnChart([]float64{}, 3, "Title", "X", "Y")
	uassert.Equal(t, "no data to display", result)
}

func TestGenerateColumnChart_Normalization(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6}
	result := GenerateColumnChart(values, 3, "Norm", "X", "Y")
	uassert.True(t, strings.Contains(result, "Norm"))
}
