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
expose them in the Realm. 

[embedmd]:# (../assets/how-to-guides/creating-grc20/mytoken-2.gno go)
```go
func TotalSupply() uint64 {
	return mytoken.TotalSupply()
}

func Decimals() uint {
	return mytoken.GetDecimals()
}

func BalanceOf(account std.Address) uint64 {
	balance, err := mytoken.BalanceOf(account)
	if err != nil {
		panic(err)
	}

	return balance
}

func Allowance(owner, spender std.Address) uint64 {
	allowance, err := mytoken.Allowance(owner, spender)
	if err != nil {
		panic(err)
	}

	return allowance
}

func Transfer(recipient std.Address, amount uint64) {
	caller := std.PrevRealm().Addr()
	if err := mytoken.Transfer(caller, recipient, amount); err != nil {
		panic(err)
	}
}

func Approve(spender std.Address, amount uint64) {
	caller := std.PrevRealm().Addr()
	if err := mytoken.Approve(caller, spender, amount); err != nil {
		panic(err)
	}
}

func TransferFrom(from, to std.Address, amount uint64) {
	caller := std.PrevRealm().Addr()

	if amount <= 0 {
		panic("transfer amount must be greater than zero")
	}

	if err := mytoken.TransferFrom(caller, from, to, amount); err != nil {
		panic(err)
	}
}

func Mint(address std.Address, amount uint64) {
	assertIsAdmin(std.PrevRealm().Addr())

	if amount <= 0 {
		panic("mint amount must be greater than zero")
	}

	if err := mytoken.Mint(address, amount); err != nil {
		panic(err)
	}
}

func Burn(address std.Address, amount uint64) {
	assertIsAdmin(std.PrevRealm().Addr())

	if amount <= 0 {
		panic("burn amount must be greater than zero")
	}

	if err := mytoken.Burn(address, amount); err != nil {
		panic(err)
	}
}

func assertIsAdmin(address std.Address) {
	if address != admin {
		panic("restricted access")
	}
}

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return mytoken.RenderHome()
	case c == 2 && parts[0] == "balance":
		owner := std.Address(parts[1])
		balance, _ := mytoken.BalanceOf(owner)
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}
```

Detailing what is happening in the above code:
- Calling the `TotalSupply` method would return the total number of tokens minted.
- Calling the `BalanceOf` method would return the total balance of an account.
- Calling the `Allowance` method would set an account as an allowed spender to
serve on behalf of the owner.
- Calling the `transfer` method transfers a configurable amount of token
from the calling account to another account, either owned or unowned.
- Calling the `Approve` method approves a calling account to spend a
configurable amount of token(s) on behalf of the token owner.
- Calling the `TransferFrom` method transfers a configurable amount of 
token from an account that granted approval to another account, either owned or unowned.
- Calling the `Mint` method creates a configurable number of tokens by 
the administrator.
- Calling the `Burn` method destroys a configurable number of tokens by
the administrator.
- Calling the `Render` method returns a user's `balance` as a formatted
string. Learn more about the `Render`
  method and how it's used [here](../concepts/realms.md).
- Lastly, we provide a local function designed to verify that the calling account is
indeed the owner; it triggers a panic if this is not the case. This critical function acts
as a safeguard to prevent unauthorized actions by non-administrators.


You can view the full code on [this Playground link](https://play.gno.land/p/1UXqufodX6f).


## Conclusion

That's it 🎉

You have successfully built a simple GRC20 Realm that is ready to be deployed on the Gno chain and called by users.
In the upcoming guides, we will see how we can develop more complex Realm logic and have them interact with outside tools like a wallet application.
