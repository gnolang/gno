package smartcontractwalletdemo

import (
	"gno.land/p/ursulovic/smartcontractwallet"
)

var (
	wallet smartcontractwallet.Wallet
)

func init() {
	owners := []string {
		"g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x",
		"g1mw2xft3eava9kfhqw3fjj3kkf3pkammty0mtv7",
		"g1sz9vzfgthtu370l8w3gyl2rg4qmz8txn2tw5vp",
	}
	wallet.InitOwners(owners, 0.5)
}

func DepositCoins() {
	wallet.DepositCoins()
}

func ProposeWithdraw(recipient string, amount int64, denom string, approvalPercentage float64) {
	wallet.ProposeWithdraw(recipient, amount, denom, approvalPercentage)
}

func ExecuteWithdraw(id int64) {
	wallet.ExecuteWithdraw(id)
}

func ApproveWithdraw(id int64) {
	wallet.ApproveWithdraw(id)
}

func Render(s string) string {
	return ""
}