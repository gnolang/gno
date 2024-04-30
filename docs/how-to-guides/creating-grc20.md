---
id: creating-grc20
---

# How to Create a GRC20 Token
## Overview

This guide shows you how to write a simple **GRC20**
a [Realm](../concepts/realms.md), in [Gno](../concepts/gno-language.md). For actually deploying the Realm, please see the
[deployment](deploy.md) guide.

Our **GRC20** Realm will have the following functionality:

- Minting a configurable amount of token.
- Keeping track of total token supply.
- Fetching the balance of an account.

## 1. Importing token package
For this realm, we import the `grc20` package, as this includes
the main functionality of our token realm. The package can be found the 
`gno.land/p/demo/grc/grc20` path.

[embedmd]:# (../assets/how-to-guides/creating-grc20/mytoken-1.gno go)
```go
package mytoken

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
)

var (
	mytoken *grc20.AdminToken
	admin   std.Address
)

// init is called once at time of deployment
func init() {
	// Set deployer of Realm to admin
	admin = std.PrevRealm().Addr()

	// Set token name, symbol and number of decimals
	mytoken = grc20.NewAdminToken("My Token", "TKN", 4)

	// Mint 1 million tokens to admin
	mytoken.Mint(admin, 1000000*10000)
}
```

The code snippet above does the following:
- Defines a new token variable, `mytoken`, and assigns it to a
pointer to the GRC20 token type, `grc20.AdminToken`,
- Defines and sets the value of `admin` with a type of `std.Address` to contain 
the address of the deployer
- Initializes `mytoken` as a new GRC20 token, and set its name, symbol, and 
decimal values,
- Mint 1 million units of `My Token` and assign them to the admin's address.

## 2. Adding token functionality

In order to call exported functions from the `grc20` package, we also need to 
expose them in the realm. Let's go through all functions in the GRC20 package,
one by one:

```go
// TotalSupply returns the total supply of mytoken
func TotalSupply() uint64 {
	return mytoken.TotalSupply()
}

```
Calling the `TotalSupply` method would return the total number of tokens minted.

```go
// Decimals returns the number of decimals of mytoken
func Decimals() uint {
	return mytoken.GetDecimals()
}
```
Calling the `Decimals` method would return number of decimals of the token.

```go
// BalanceOf returns the balance mytoken for `account`
func BalanceOf(account std.Address) uint64 {
	balance, err := mytoken.BalanceOf(account)
	if err != nil {
		panic(err)
	}

	return balance
}
```

Calling the `BalanceOf` method would return the total balance of an account.

```go
// Allowance returns the allowance of spender on owner's balance
func Allowance(owner, spender std.Address) uint64 {
	allowance, err := mytoken.Allowance(owner, spender)
	if err != nil {
		panic(err)
	}

	return allowance
}
```
Calling the `Allowance` method will return the amount `spender` is allowed to spend
from `owner`'s balance.

```go
// Transfer transfers amount from caller to recipient
func Transfer(recipient std.Address, amount uint64) {
	caller := std.PrevRealm().Addr()
	if err := mytoken.Transfer(caller, recipient, amount); err != nil {
		panic(err)
	}
}
```
Calling the `Transfer` method transfers amount of token from the calling account 
to the recipient account.

```go
func Approve(spender std.Address, amount uint64) {
	caller := std.PrevRealm().Addr()
	if err := mytoken.Approve(caller, spender, amount); err != nil {
		panic(err)
	}
}
```
Calling the `Approve` method approves `spender` to spend `amount` from the caller's
balance of tokens.

```go
// TransferFrom transfers `amount` of tokens from `from` to `to` 
func TransferFrom(sender, recipient std.Address, amount uint64) {
	caller := std.PrevRealm().Addr()

	if amount <= 0 {
		panic("transfer amount must be greater than zero")
	}

	if err := mytoken.TransferFrom(caller, from, to, amount); err != nil {
		panic(err)
	}
}
```
Calling the `TransferFrom` method moves `amount` of tokens from `sender` to 
`recipient` using the allowance mechanism. `amount` is then deducted from the
caller’s allowance.

```go
// Mint mints amount of tokens to address. Callable only by admin of token
func Mint(address std.Address, amount uint64) {
	assertIsAdmin(std.PrevRealm().Addr())

	if amount <= 0 {
		panic("mint amount must be greater than zero")
	}

	if err := mytoken.Mint(address, amount); err != nil {
		panic(err)
	}
}
```
Calling the `Mint` method creates `amount` of tokens and assigns them to `address`,
increasing the total supply.

```go
// Burn burns amount of tokens from address. Callable only by admin of token
func Burn(address std.Address, amount uint64) {
	assertIsAdmin(std.PrevRealm().Addr())

	if amount <= 0 {
		panic("burn amount must be greater than zero")
	}

	if err := mytoken.Burn(address, amount); err != nil {
		panic(err)
	}
}
```
Calling the `Mint` method burns `amount` of tokens from the balance of `address`,
decreasing the total supply.

```go
// assertIsAdmin asserts the address is the admin of token
func assertIsAdmin(address std.Address) {
	if address != admin {
		panic("restricted access")
	}
}
```
Calling the `assertIsAdmin` method checks if `address` is equal to the 
package-level `admin` variable. 

```go
// Render renders the state of the realm
func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		// Default GRC20 render
		return mytoken.RenderHome()
	case c == 2 && parts[0] == "balance":
		// Render balance of specific address
		owner := std.Address(parts[1])
		balance, _ := mytoken.BalanceOf(owner)
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}
```
Calling the `Render` method returns a general render of the GRC20 realm, or
if given a specific address, the user's `balance` as a formatted string.

You can view the full code on [this Playground link](https://play.gno.land/p/km7Ja6WDQoL).

## Conclusion
That's it 🎉

You have successfully built a simple GRC20 Realm that is ready to be deployed on the Gno chain and called by users.
In the upcoming guides, we will see how we can develop more complex Realm logic and have them interact with outside tools like a wallet application.
