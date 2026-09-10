package memberstore

import (
	"strings"

	"gno.land/p/samcrew/piechart"
)

// RenderCharts generates two pie charts for member tiers:
// 1) distribution of member counts per tier
// 2) distribution of power per tier
func RenderCharts(members MembersByTier) string {
	var sb strings.Builder

	tierNames := []string{T1, T2, T3}
	pieSlicesTs := make([]piechart.PieSlice, 0, len(tierNames))
	pieSlicesTp := make([]piechart.PieSlice, 0, len(tierNames))

	for _, tn := range tierNames {
		tier, ok := tiers.GetTier(tn)
		if !ok {
			return ""
		}

		ts := float64(members.GetTierSize(tn))
		tp := tier.PowerHandler(members, tiers) * ts

		pieSlicesTs = append(pieSlicesTs, piechart.PieSlice{
			Value: ts,
			Color: tierColor(tn),
			Label: tn,
		})
		pieSlicesTp = append(pieSlicesTp, piechart.PieSlice{
			Value: tp,
			Color: tierColor(tn),
			Label: tn,
		})
	}

	// Render pie charts for members count and power distribution
	resultPieChartTs := piechart.Render(pieSlicesTs, "Members distribution:")
	resultPieChartTp := piechart.Render(pieSlicesTp, "Power distribution:")

	sb.WriteString(resultPieChartTs + "\n")
	sb.WriteString(resultPieChartTp + "\n")

	return sb.String()
}
