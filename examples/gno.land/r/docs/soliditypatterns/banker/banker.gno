package banker

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"strconv"

	"gno.land/p/nt/avl"
)

// This example demonstrates the Gno equivalent of Solidity's
// payable functions and ETH transfers.

var deposits avl.Tree // address -> int64

func init() {
	deposits = avl.Tree{}
}

// Deposit accepts coins and tracks the sender's balance
// In Solidity, this would be a payable function
func Deposit(_ realm) {
	caller := runtime.PreviousRealm().Address()
	sent := banker.OriginSend()

	// Check if any coins were sent
	if sent.AmountOf("ugnot") <= 0 {
		panic("must send some ugnot")
	}

	amount := sent.AmountOf("ugnot")

	// Get current balance
	currentBalance := int64(0)
	if val, exists := deposits.Get(caller.String()); exists {
		currentBalance = val.(int64)
	}

	// Update balance
	deposits.Set(caller.String(), currentBalance+amount)
}

// Withdraw sends coins back to the caller
// This is equivalent to Solidity's transfer() or call{value: x}()
func Withdraw(_ realm, amount int64) {
	caller := runtime.PreviousRealm().Address()

	// Get current balance
	currentBalance := int64(0)
	if val, exists := deposits.Get(caller.String()); exists {
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
	bnk := banker.NewBanker(banker.BankerTypeRealmSend)

	// Get the realm's address
	realmAddr := runtime.CurrentRealm().Address()

	// Send coins from realm to caller
	bnk.SendCoins(realmAddr, caller, chain.Coins{chain.NewCoin("ugnot", amount)})
}

// GetBalance returns the deposited balance for an address
func GetBalance(addr address) int64 {
	if val, exists := deposits.Get(addr.String()); exists {
		return val.(int64)
	}
	return 0
}

// GetRealmBalance returns the total coins held by this realm
func GetRealmBalance() int64 {
	bnk := banker.NewBanker(banker.BankerTypeReadonly)
	realmAddr := runtime.CurrentRealm().Address()
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
import (
    "chain/banker"
    "chain/runtime"
)

func Deposit(_ realm) {
    caller := runtime.PreviousRealm().Address()
    sent := banker.OriginSend()
    
    // Unlike Solidity's single msg.value, Gno supports multiple coin types
    // A user can send ugnot, custom tokens, or both in one transaction
    amount := sent.AmountOf("ugnot")
    deposits[caller] += amount
}
` + "```" + `

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
    "chain/runtime"
)

func Withdraw(_ realm, amount int64) {
    caller := runtime.PreviousRealm().Address()
    
    // Update state BEFORE sending (prevents reentrancy)
    deposits[caller] -= amount
    
    // Create banker with send capability
    bnk := banker.NewBanker(banker.BankerTypeRealmSend)
    
    // Send coins
    realmAddr := runtime.CurrentRealm().Address()
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
