package reentrancy

// This example demonstrates why reentrancy attacks are NOT possible in Gno,
// unlike in Solidity where they have caused major vulnerabilities.

func Render(_ string) string {
	return `# Reentrancy in Gno

## The Problem in Solidity

Reentrancy is one of the most notorious vulnerabilities in Solidity, responsible for 
the infamous DAO hack that led to a $60M loss and Ethereum's hard fork.

### How Reentrancy Works in Solidity

` + "```solidity" + `
// VULNERABLE Solidity contract
contract VulnerableVault {
    mapping(address => uint) public balances;
    
    function withdraw() public {
        uint amount = balances[msg.sender];
        
        // DANGER: External call BEFORE state update
        (bool success, ) = msg.sender.call{value: amount}("");
        require(success);
        
        // State updated AFTER external call
        balances[msg.sender] = 0;
    }
}

// Attacker contract
contract Attacker {
    VulnerableVault vault;
    
    // This function is called when receiving ETH
    receive() external payable {
        // Re-enter the withdraw function!
        if (address(vault).balance >= 1 ether) {
            vault.withdraw();
        }
    }
}
` + "```" + `

The attack flow:
1. Attacker deposits 1 ETH
2. Attacker calls withdraw()
3. Vault sends ETH to attacker
4. Attacker's receive() is triggered
5. Attacker calls withdraw() AGAIN (balance not yet updated!)
6. Repeat until vault is drained

## Why Reentrancy is NOT Possible in Gno

Gno's architecture fundamentally prevents reentrancy. 

In Solidity, when you send ETH, the recipient's ` + "`receive()`" + ` or ` + "`fallback()`" + ` 
function is automatically called. This creates the reentrancy vector.

**In Gno, coin transfers do not trigger any code execution on the recipient.**
The only way to execute code in another realm is through an explicit function call
using the ` + "`cross`" + ` keyword, making all realm boundary crossings visible and intentional.
See [Crossing](/r/docs/crossing) for more details.

` + "```gno" + `
// This is SAFE in Gno
func Withdraw(_ realm, amount int64) {
    caller := runtime.PreviousRealm().Address()
    
    // Even if we update state AFTER sending...
    bnk := banker.NewBanker(banker.BankerTypeRealmSend)
    bnk.SendCoins(realmAddr, caller, coins)
    
    // No code executes on the recipient!
    // The caller cannot re-enter this function
    deposits[caller] = 0
}
` + "```" + `

Gno transactions are atomic. If any part fails, the entire transaction reverts.
This means you can't have a partially-executed state that an attacker could exploit.

## Best Practices

Even though reentrancy isn't possible in Gno, following the
**Checks-Effects-Interactions** pattern keeps code readable and maintainable:

` + "```gno" + `
func Withdraw(_ realm, amount int64) {
    caller := runtime.PreviousRealm().Address()
    
    // 1. CHECKS - Validate all conditions
    if deposits[caller] < amount {
        panic("insufficient balance")
    }
    
    // 2. EFFECTS - Update state
    deposits[caller] -= amount
    
    // 3. INTERACTIONS - External calls last
    bnk := banker.NewBanker(banker.BankerTypeRealmSend)
    bnk.SendCoins(realmAddr, caller, coins)
}
` + "```" + `

---

[Back to Solidity Patterns](/r/docs/soliditypatterns) | [Banker Example](/r/docs/soliditypatterns/banker)
`
}
