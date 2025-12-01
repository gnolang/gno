package subscriptions

import (
	"chain"
	"time"
)

// Helper to create through MsgCall with GNOT native token price and monthly renewal period
func NewServiceWithMonthlyRenewal(cur realm, displayName, description string, price uint64) {
	coins := chain.NewCoins(chain.NewCoin("ugnot", int64(price)))
	NewService(cur, displayName, description, 10*time.Second, coins, chain.Coins{})
}
