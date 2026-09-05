package payrolls

import (
	"chain/banker"
	"testing"
)

func BalanceOf(t *testing.T, addr address, denom string) int64 {
	t.Helper()

	banker_ := banker.NewBanker(banker.BankerTypeReadonly)
	return banker_.GetCoins(addr).AmountOf(denom)
}
