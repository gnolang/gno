package mytoken

import "strings"

func Render(path string) string {
	content := `
# My Token

This Solidity contract defines a basic ERC-20 token with a fixed supply, all tokens initially assigned to the deployer.

---

### Solidity Version

^^^solidity
contract MyToken is ERC20 {
    constructor(uint256 initialSupply) ERC20("MyToken", "MTK") {
        _mint(msg.sender, initialSupply * 10 ** decimals());
    }
}
^^^

#### Explanation

This Solidity contract does:

* Declare an ERC-20 token contract named ^MyToken^ with symbol ^MTK^.
* In the constructor, mint the ^initialSupply^ of tokens to the deployer (^msg.sender^).
* Inherits from ^ERC20^, so it automatically has the standard ERC-20 functions:
	* ^name()^ → returns ^"MyToken"^
	* ^symbol()^ → returns ^"MTK"^
	* ^decimals()^ → returns ^18^
	* ^totalSupply()^ → returns total tokens
	* ^balanceOf(address)^ → checks token balance
	* ^transfer(to, amount)^ → sends tokens
	* ^approve(spender, amount)^ → approves spending
	* ^allowance(owner, spender)^ → checks allowance
	* ^transferFrom(from, to, amount)^ → transfers tokens on behalf of another account

---

### Gno Version

^^^go
package mytoken

import (
	"std"

	"gno.land/p/demo/grc/grc20"
)

var (
	token *grc20.Token
	ledger *grc20.PrivateLedger
)

func init() {
	token, ledger = grc20.NewToken("MyToken", "MTK", 6)

	admin := std.CurrentRealm().Address()
	_ = ledger.Mint(admin, 1_000_000*1_000_000)
}

func Name() string {
    return token.GetName()
}

func Symbol() string {
    return token.GetSymbol()
}

func Decimals() int {
    return token.GetDecimals()
}

func TotalSupply() int64 {
    return token.TotalSupply()
}

func BalanceOf(addr std.Address) int64 {
    return token.BalanceOf(addr)
}

func Transfer(to std.Address, amount int64) error {
    teller := token.CallerTeller()
    return teller.Transfer(to, amount)
}

func Approve(spender std.Address, amount int64) error {
    teller := token.CallerTeller()
    return teller.Approve(spender, amount)
}

func TransferFrom(owner, to std.Address, amount int64) error {
    teller := token.CallerTeller()
    return teller.TransferFrom(owner, to, amount)
}
^^^

* The contract creates a token ^MyToken^ with symbol ^MTK^ and 18 decimals.
* All tokens are minted initially to the admin (current realm address) in the ^init()^ function.
* Standard ERC-20-like methods are provided:
	* ^Name()^, ^Symbol()^, ^Decimals()^, ^TotalSupply()^ → token info
	* ^BalanceOf(addr)^ → check account balance
	* ^Transfer(to, amount)^ → transfer tokens
	* ^Approve(spender, amount)^ → allow another address to spend tokens
	* ^TransferFrom(owner, to, amount)^ → transfer tokens on behalf of someone

---
`
	return strings.ReplaceAll(content+RenderDemo(), "^", "`")
}
