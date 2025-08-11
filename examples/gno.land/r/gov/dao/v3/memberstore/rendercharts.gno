package memberstore

import (
	"strings"

	"gno.land/p/samcrew/piechart"
)

// RenderCharts generates two pie charts for member tiers:
// 1) distribution of member counts per tier
// 2) distribution of power per tier
// It writes into the provided string builder
func RenderCharts(members MembersByTier, sb *strings.Builder) string {
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

	colors := []string{"#2FA08C", "#8654D6", "#F5C65B"}

	resultPieTs := piechart.GeneratePieChart(memberCounts, colors, tierNames, "Members repartition:")
	resultPieTp := piechart.GeneratePieChart(powerValues, colors, tierNames, "Power repartition:")

	sb.WriteString(resultPieTs)
	sb.WriteString(resultPieTp)
	sb.WriteString("\n")

	return sb.String()
}
