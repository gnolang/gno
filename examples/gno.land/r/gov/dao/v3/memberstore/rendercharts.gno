package memberstore

import (
	"strings"

	"gno.land/p/samcrew/piechart"
	"gno.land/p/sunspirit/md"
)

// RenderCharts generates two pie charts for member tiers:
// 1) distribution of member counts per tier
// 2) distribution of power per tier
func RenderCharts(members MembersByTier) string {
	var sb strings.Builder

	tierNames := []string{"T1", "T2", "T3"}
	memberCounts := make([]float64, 0, len(tierNames))
	powerValues := make([]float64, 0, len(tierNames))

	for _, tn := range tierNames {
		tier, ok := Tiers.GetTier(tn)
		if !ok {
			return ""
		}
		ts := float64(members.GetTierSize(tn))
		memberCounts = append(memberCounts, ts)
		tp := tier.PowerHandler(members, Tiers) * ts
		powerValues = append(powerValues, tp)
	}

	colors := []string{"#7CB342", "#AB47BC", "#FFCC32"}

	resultPieTs := piechart.RenderPieChart(memberCounts, colors, tierNames, "Members repartition:")
	resultPieTp := piechart.RenderPieChart(powerValues, colors, tierNames, "Power repartition:")

	sb.WriteString(resultPieTs)
	sb.WriteString(md.LineBreak(1))
	sb.WriteString(resultPieTp)
	sb.WriteString(md.LineBreak(1))

	return sb.String()
}
