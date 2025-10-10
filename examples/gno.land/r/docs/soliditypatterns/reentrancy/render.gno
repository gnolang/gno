package reentrancy

import "strings"

func Render(path string) string {
	content := `
# Reentrancy Guard

Prevents reentrancy attacks by blocking recursive calls during execution.

---

## Solidity Version

^^^solidity
contract ReentrancyGuard {
    bool private _locked;

    modifier nonReentrant() {
        require(!_locked, "ReentrancyGuard: reentrant call");
        _locked = true;
        _;
        _locked = false;
    }

    mapping(address => uint256) public balances;

    function deposit() public payable {
        balances[msg.sender] += msg.value;
    }

    function withdraw(uint256 amount) public nonReentrant {
        require(balances[msg.sender] >= amount, "Insufficient balance");

        balances[msg.sender] -= amount;
        (bool success, ) = msg.sender.call{value: amount}("");
        require(success, "Transfer failed");
    }
}
^^^

#### Explanation

This Solidity implementation includes:

* A ^_locked^ boolean to track if a function is currently executing
* ^nonReentrant^ modifier that sets the lock before execution and clears it after
* ^balances^ mapping to track user deposits in the contract
* ^deposit()^ function allows users to send ETH and updates their balance
* ^withdraw()^ function uses the ^nonReentrant^ modifier to prevent reentrancy attacks
* The modifier ensures ^_locked^ is false, sets it to true, executes the function, then resets to false
* Critical operations (balance update and external call) happen in the correct order

---

## Gno Version

^^^go
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

    balances[caller] = std.NewCoins(std.NewCoin("ugnot", newBalance))

    banker := std.NewBanker(std.BankerTypeRealmSend)
    banker.SendCoins(std.CurrentRealm().Address(), caller, amount)
}
^^^

#### Key Differences

* Gno uses ^defer^ to ensure unlock happens even if panic occurs
* ^std.Coins^ operations for multi-token support
* Manual lock/unlock instead of modifier pattern
* ^std.OriginCaller()^ gets the transaction originator
* Balance operations use ^AmountOf()^ for specific denominations

---
`
	return strings.ReplaceAll(content+RenderDemo(), "^", "`")
}
