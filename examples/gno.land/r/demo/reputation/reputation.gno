package reputation

import (
	"strconv"
	"strings"

	"gno.land/p/nt/avl/v0"
)

var (
	scores avl.Tree
	totals avl.Tree
)

func AddPoints(cur realm, target address, category string, points int64) {
	caller := cur.Previous()
	if caller.IsUserCall() {
		panic("reputation: AddPoints can only be called by another realm, not directly by a user")
	}
	if points <= 0 {
		panic("reputation: points must be positive")
	}
	issuer := caller.PkgPath()

	key := scoreKey(target, category, issuer)
	current := int64(0)
	if raw := scores.Get(key); raw != nil {
		current = raw.(int64)
	}
	scores.Set(key, current+points)

	totalKey := target.String()
	currentTotal := int64(0)
	if raw := totals.Get(totalKey); raw != nil {
		currentTotal = raw.(int64)
	}
	totals.Set(totalKey, currentTotal+points)
}

func GetScore(target address, category string, issuerPkgPath string) int64 {
	raw := scores.Get(scoreKey(target, category, issuerPkgPath))
	if raw == nil {
		return 0
	}
	return raw.(int64)
}

func GetTotalScore(target address) int64 {
	raw := totals.Get(target.String())
	if raw == nil {
		return 0
	}
	return raw.(int64)
}

func scoreKey(target address, category string, issuerPkgPath string) string {
	return target.String() + "|" + category + "|" + issuerPkgPath
}

func Render(path string) string {
	var sb strings.Builder
	sb.WriteString("# Reputation Ledger\n\n")
	sb.WriteString("Generic, cross-realm reputation scores. Points can only be awarded by other realms (not directly by users), and every entry is attributed to the realm that issued it.\n\n")
	sb.WriteString("## Total scores\n\n")
	totals.Iterate("", "", func(key string, value interface{}) bool {
		sb.WriteString("- " + key + ": " + strconv.FormatInt(value.(int64), 10) + "\n")
		return false
	})
	return sb.String()
}
