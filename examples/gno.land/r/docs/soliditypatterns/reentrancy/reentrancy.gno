package reentrancy

func Render(_ string) string {
	return `# Reentrancy in Gno

In Solidity, sending ETH automatically triggers the recipient's ` + "`receive()`" + ` or ` + "`fallback()`" + ` 
function. An attacker can exploit this to call back into the original contract before state 
updates complete—the infamous DAO hack exploited this vulnerability for a $60M loss.

## The Solidity Problem

` + "```solidity" + `
contract VulnerableVault {
    mapping(address => uint) balances;
    
    function withdraw() public {
        uint amount = balances[msg.sender];
        msg.sender.call{value: amount}("");  // 1. Sends ETH, triggering attacker's receive()
        balances[msg.sender] = 0;            // 3. State update happens too late!
    }
}

contract Attacker {
    VulnerableVault vault;
    
    receive() external payable {
        if (address(vault).balance >= 1 ether) {
            vault.withdraw();  // 2. Re-enters withdraw() while balance is still non-zero
        }
    }
}
` + "```" + `

## Why Gno is Safe

Coin transfers in Gno do **not** execute code on the recipient. Unlike Solidity's 
` + "`receive()`" + ` and ` + "`fallback()`" + ` hooks, Gno realm code only runs through explicit 
` + "`cross`" + ` function calls—eliminating the reentrancy attack vector entirely.

` + "```go" + `
func Withdraw(_ realm, amount int64) {
    caller := std.PreviousRealm().Address()
    
    bnk := std.NewBanker(std.BankerTypeRealmSend)
    coins := std.NewCoins(std.NewCoin("ugnot", amount))
    bnk.SendCoins(std.CurrentRealm().Address(), caller, coins)  // No code executes on recipient
    deposits[caller] = 0  // Safe - no reentrancy possible
}
` + "```" + `

## Best Practice: Checks-Effects-Interactions

Even though Gno prevents reentrancy attacks, following the **Checks-Effects-Interactions** 
pattern keeps your code clear and maintainable:

` + "```go" + `
func Withdraw(_ realm, amount int64) {
    caller := std.PreviousRealm().Address()
    
    // 1. Checks - validate all preconditions
    balance := deposits[caller]
    if balance < amount {
        panic("insufficient balance")
    }
    
    // 2. Effects - update state before external calls
    deposits[caller] -= amount
    
    // 3. Interactions - perform external operations last
    bnk := std.NewBanker(std.BankerTypeRealmSend)
    coins := std.NewCoins(std.NewCoin("ugnot", amount))
    bnk.SendCoins(std.CurrentRealm().Address(), caller, coins)
}
` + "```" + `

This pattern remains a best practice for code clarity and future-proofing, even when 
reentrancy isn't a concern.

---


## See Also

- [Banker](/r/docs/banker) - Coin transfer patterns
`
}
