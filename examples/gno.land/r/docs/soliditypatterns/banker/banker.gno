// Package banker recreates Solidity's payable/withdraw pattern in Gno.
// It accepts deposits and lets callers withdraw their balance through the chain banker.
package banker

import (
	"chain"
	"chain/banker"
	"chain/runtime/unsafe"
	"strconv"

	"gno.land/p/nt/avl/v0"
)

// This example demonstrates the Gno equivalent of Solidity's
// payable functions and ETH transfers.

var deposits avl.Tree // address -> int64

// Deposit accepts coins and tracks the sender's balance
// In Solidity, this would be a payable function
func Deposit(cur realm) {
	// Require a direct user call so coins are credited to the real sender.
	if !cur.Previous().IsUserCall() {
		panic("invalid non-user call")
	}

	caller := cur.Previous().Address()
	sent := unsafe.OriginSend()

	// Check that coins were sent
	amount := sent.AmountOf("ugnot")
	if amount <= 0 {
		panic("must send some ugnot")
	}

	// Get current balance
	currentBalance := int64(0)
	if val := deposits.Get(caller.String()); val != nil {
		currentBalance = val.(int64)
	}

	// Update balance
	deposits.Set(caller.String(), currentBalance+amount)
}

// Withdraw sends coins back to the caller
// This is equivalent to Solidity's transfer() or call{value: x}()
func Withdraw(cur realm, amount int64) {
	// Require a direct user call so only the balance owner can withdraw.
	if !cur.Previous().IsUserCall() {
		panic("invalid non-user call")
	}

	// A negative amount would inflate the balance and send negative coins.
	if amount <= 0 {
		panic("amount must be positive")
	}

	caller := cur.Previous().Address()

	// Get current balance
	currentBalance := int64(0)
	if val := deposits.Get(caller.String()); val != nil {
		currentBalance = val.(int64)
	}

	// Check balance
	if currentBalance < amount {
		panic("insufficient balance")
	}

	newBalance := currentBalance - amount
	if newBalance == 0 {
		deposits.Remove(caller.String())
	} else {
		deposits.Set(caller.String(), newBalance)
	}

	// Create a banker to send coins
	bnk := banker.NewBanker(banker.BankerTypeRealmSend, cur)

	// Get the realm's address
	realmAddr := cur.Address()

	// Send coins from realm to caller
	bnk.SendCoins(realmAddr, caller, chain.Coins{chain.NewCoin("ugnot", amount)})
}

// GetBalance returns the deposited balance for an address
func GetBalance(addr address) int64 {
	balance := deposits.Get(addr.String())
	if balance == nil {
		return 0
	}
	return balance.(int64)
}

// GetRealmBalance returns the total coins held by this realm
func GetRealmBalance() int64 {
	bnk := banker.NewReadonlyBanker()
	realmAddr := unsafe.CurrentRealm().Address()
	coins := bnk.GetCoins(realmAddr)
	return coins.AmountOf("ugnot")
}

func RenderDemo() string {
	out := "## Live Demo\n\n"
	out += "Realm Balance: **" + strconv.FormatInt(GetRealmBalance(), 10) + " ugnot**\n\n"
	out += "[Deposit](/r/docs/soliditypatterns/banker$help&func=Deposit) | "
	out += "[Withdraw](/r/docs/soliditypatterns/banker$help&func=Withdraw)\n"
	return out
}

func Render(_ string) string {
	return `# Sending Coins (Banker)

In Solidity, you send ETH using ` + "`transfer()`" + `, ` + "`send()`" + `, or ` + "`call{value: x}()`" + `.
In Gno, you use the **Banker** to send coins.

` + RenderDemo() + `
## Solidity vs Gno Comparison

### Solidity: Receiving ETH

` + "```solidity" + `
// Payable function receives ETH
function deposit() public payable {
    balances[msg.sender] += msg.value;
}
` + "```" + `

### Gno: Receiving Coins

` + "```go" + `
import "chain/runtime/unsafe"

var deposits map[address]int64

func Deposit(cur realm) {
    // Require a direct user call so coins are credited to the real sender.
    if !cur.Previous().IsUserCall() {
        panic("invalid non-user call")
    }
    caller := cur.Previous().Address()
    sent := unsafe.OriginSend()

    // Unlike Solidity's single msg.value, Gno supports multiple coin types
    // A user can send ugnot, custom tokens, or both in one transaction
    amount := sent.AmountOf("ugnot")
    deposits[caller] += amount
}
` + "```" + `

> **Note:** These snippets use a map for readability. The live source stores balances in
> an AVL tree (` + "`avl.Tree`" + `) for deterministic iteration and performance. See [source code](/r/docs/soliditypatterns/banker$source).

---

### Solidity: Sending ETH

` + "```solidity" + `
// Using transfer (reverts on failure)
payable(recipient).transfer(amount);

// Using call
(bool success, ) = recipient.call{value: amount}("");
require(success, "Transfer failed");
` + "```" + `

### Gno: Sending Coins

` + "```go" + `
import (
    "chain"
    "chain/banker"
    "chain/runtime/unsafe"
)

var deposits map[address]int64

func Withdraw(cur realm, amount int64) {
    // Require a direct user call so only the balance owner can withdraw.
    if !cur.Previous().IsUserCall() {
        panic("invalid non-user call")
    }
    caller := cur.Previous().Address()

    // Checks: reject non-positive amounts and over-withdrawal.
    if amount <= 0 {
        panic("amount must be positive")
    }
    if deposits[caller] < amount {
        panic("insufficient balance")
    }

    // Effects: update state before sending (checks-effects-interactions)
    deposits[caller] -= amount

    // Interactions: create banker with send capability and send coins
    bnk := banker.NewBanker(banker.BankerTypeRealmSend, cur)
    realmAddr := cur.Address()
    bnk.SendCoins(realmAddr, caller, chain.Coins{chain.NewCoin("ugnot", amount)})
}
` + "```" + `

## Banker Types

| Type | Purpose |
|------|---------|
| ` + "`BankerTypeReadonly`" + ` | Read balances only |
| ` + "`BankerTypeRealmSend`" + ` | Send coins from realm |
| ` + "`BankerTypeRealmIssue`" + ` | Issue new realm-specific tokens |

## Key Differences from Solidity

Gno requires an explicit Banker with appropriate permissions.
Reentrancy attacks are not possible (see [Reentrancy](/r/docs/soliditypatterns/reentrancy)).
Multiple coin denominations are supported, not just one native token.

[View source code](/r/docs/soliditypatterns/banker$source)

---

[Back to Solidity Patterns](/r/docs/soliditypatterns) | [Banker Documentation](https://docs.gno.land/resources/gno-stdlibs/#banker)
`
}
