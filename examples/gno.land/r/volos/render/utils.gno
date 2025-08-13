package render

import (
	"strconv"
	"strings"
	"time"
	"gno.land/p/demo/ufmt"
	volos "gno.land/r/volos/core"
)

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func calculateUtilization(marketId string) string {
	totalSupply := parseFloat(volos.GetMarketTotalSupplyAssets(marketId))
	totalBorrow := parseFloat(volos.GetMarketTotalBorrowAssets(marketId))
	
	if totalSupply == 0 {
		return "0.00"
	}
	
	utilization := (totalBorrow / totalSupply) * 100
	return strconv.FormatFloat(utilization, 'f', 2, 64)
}

func formatRate(rateStr string) string {
	rate := parseFloat(rateStr)
	percentage := rate * 100
	return strconv.FormatFloat(percentage, 'f', 4, 64)
}

func formatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format("02 Jan 2006, 15:04:05")
}

func formatMoney(amount float64) string {
	s := ufmt.Sprintf("%.2f", amount)
	parts := strings.Split(s, ".")
	intPart := parts[0]

	n := len(intPart)
	if n <= 3 {
		return s
	}

	out := ""
	count := 0
	for i := n - 1; i >= 0; i-- {
		if count > 0 && count%3 == 0 {
			out = "," + out
		}
		out = string(intPart[i]) + out
		count++
	}

	if len(parts) > 1 {
		out = out + "." + parts[1]
	}

	return out
}
