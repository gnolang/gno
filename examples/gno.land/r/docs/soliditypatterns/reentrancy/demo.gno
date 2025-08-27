package reentrancy

import "std"

var (
    balances map[std.Address]std.Coins
    locked   bool
)

func init() {
    balances = make(map[std.Address]std.Coins)
}

func requireNonReentrant() {
    if locked {
        panic("ReentrancyGuard: reentrant call")
    }
    locked = true
}

func unlockReentrancy() {
    locked = false
}

func Deposit() {
    std.AssertOriginCall()
    caller := std.OriginCaller()
    amount := std.OriginSend()

    balances[caller] = balances[caller].Add(amount)
}

func Withdraw(amount std.Coins) {
    std.AssertOriginCall()
    requireNonReentrant()
    defer unlockReentrancy()

    caller := std.OriginCaller()

    newBalance := balances[caller].AmountOf("ugnot") - amount.AmountOf("ugnot")
    if newBalance < 0 {
        panic("Insufficient balance")
    }

    balances[caller].Add(std.NewCoins(std.NewCoin("ugnot", newBalance)))

    banker := std.NewBanker(std.BankerTypeRealmSend)
    banker.SendCoins(std.CurrentRealm().Address(), caller, amount)
}

func IsLocked() bool {
    return locked
}

func RenderDemo() string {
    content := "## Live Demo\n\n"
    content += "**Contract Status:** "
    if locked {
        content += "🔒 Locked (call in progress)\n\n"
    } else {
        content += "🔓 Unlocked\n\n"
    }

    content += "[Deposit 100ugnot](reentrancy$help&func=Deposit&send=100ugnot)\n\n"
    content += "[Withdraw 50ugnot](reentrancy$help&func=Withdraw&args=50ugnot)\n\n"

    content += "**Security Note:** This contract uses reentrancy protection to prevent malicious recursive calls during withdrawals.\n\n"

    return content
}
