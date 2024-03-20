---
id: creating-grc20
---

# How to create a GRC20 Token
## Overview

This guide shows you how to write a simple **GRC20** Smart Contract, or rather
a [Realm](../concepts/realms.md), in [Gno](../concepts/gno-language.md). For actually deploying the Realm, please see the
[deployment](deploy.md) guide.

Our **GRC20** Realm will have the following functionality:

- Minting a configurable amount of token.
- Keeping track of total token supply.
- Fetching the balance of an account.

## Prerequisites

- **Internet connection**

## 1. Importing token package
For this realm, we'll want to import the `grc20` package as this will include
the main functionality of our token factory realm.

[embedmd]:# (../assets/how-to-guides/creating-grc20/mytoken-1.gno go)
```go
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

In this code preview, we have:
- Defined a new local variable `mytoken` and assigned that the type of
pointer to `grc20.AdminToken`,
- Defined and set the value of local variable `admin` to point to a specific
address of type `std.Address`,
- Initialize `mytoken` as a new GRC20 token, and set its name, symbol, and 
decimal values,
- Minted 1 million units of `My Token` and to the admin's address.

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
	if err := mytoken.TransferFrom(caller, from, to, amount); err != nil {
		panic(err)
	}
}

func Mint(address std.Address, amount uint64) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)

	if err := mytoken.Mint(address, amount); err != nil {
		panic(err)
	}
}

func Burn(address std.Address, amount uint64) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)

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
- Calling the `transfer` method would transfer a configurable amount of token 
from the calling account to another account, either owned or unowned.
- Calling the `Approve` method would approve a calling account to spend a
configurable amount of token on behalf of the token owner.
- Calling the `TransferFrom` method would transfer a configurable amount of 
token from an account that granted approval to another account, either owned or unowned.
- Calling the `Mint` method would create a configurable number of tokens by 
the administrator.
- Calling the `Burn` method would destroy a configurable number of tokens by
the administrator.
- Calling the `Render` method would return a user's `balance` as a formatted
string. Learn more about the `Render`
  method and how it's used [here](../concepts/realms.md).
- Finally, we provide a local function to assert that the calling account is in
fact the owner, otherwise panic. This is a very important function that serves
to prevent abuse by non-administrators.

## Conclusion

That's it 🎉

You have successfully built a simple GRC20 Realm that is ready to be deployed on the Gno chain and called by users.
In the upcoming guides, we will see how we can develop more complex Realm logic and have them interact with outside tools like a wallet application.
