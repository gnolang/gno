---
id: creating-grc20
---

# How to create a GRC20 Token

## Overview

This guide shows you how to write a simple _GRC20_ Smart Contract, or rather a [Realm](../concepts/realms.md), in [Gno (Gno)](../concepts/gno-language.md). For actually deploying the Realm, please see the [deployment](deploy.md) guide.

Our _GRC20_ Realm will have the following functionality:

- Minting a configurable amount of token.
- Keeping track of total token supply.
- Fetching the balance of an account.

## Prerequisites

We will proceed using the typical directory structure for a Realm found within the [simple-contract guide](simple-contract.md). It is also worthwhile to consult the [GRC20 interface](https://github.com/gnolang/gno/blob/master/examples/gno.land/p/demo/grc/grc20/igrc20.gno) which we will be importing and utilizing within this guide.

## 1. Importing token package
For this realm, we'll want to import the `grc20` package as this will include the main functionality of our token factory realm.

[embedmd]:# (../assets/how-to-guides/creating-grc20/mytoken-1.gno go)
```go
package mytoken

import (
	"std"

	"gno.land/p/demo/grc/grc20"
)

var (
	mytoken *grc20.AdminToken
	admin   std.Address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // set admin account
)

// init is a constructor function that runs only once (at time of deployment)
func init() {
	// provision the token's name, symbol and number of decimals
	mytoken = grc20.NewAdminToken("Mytoken", "MTKN", 4)

	// set the total supply
	mytoken.Mint(admin, 1000000*10000) // @administrator (supply = 1 million)
}
```

In this code preview, we have:
- Defined a new local variable `mytoken` and assigned that the type of pointer to `grc20.AdminToken`.
- Defined and set the value of local variable `admin` to point to a specific gno.land address of type `std.Address`.
- Set the value of `mytoken` (type `*AdminToken`) to equal the result of creating a new token and configuring its name, symbol + decimal representation.
- Minted 1 million `Mytoken` and set the administrator as the owner of these tokens.

## 2. Adding token functionality

The following section will be about introducing Public functions to expose functionality imported from the [grc20 package](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo/grc/grc20).

Add this imports :

```go
import (
	....
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/users"

	pusers "gno.land/p/demo/users"
)
```

[embedmd]:# (../assets/how-to-guides/creating-grc20/mytoken-2.gno go)
```go
func TotalSupply() uint64 {
	return mytoken.TotalSupply()
}

func BalanceOf(owner pusers.AddressOrName) uint64 {
	balance, err := mytoken.BalanceOf(users.Resolve(owner))
	if err != nil {
		panic(err)
	}
	return balance
}

func Allowance(owner, spender pusers.AddressOrName) uint64 {
	allowance, err := mytoken.Allowance(users.Resolve(owner), users.Resolve(spender))
	if err != nil {
		panic(err)
	}
	return allowance
}

func Transfer(to pusers.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	err := mytoken.Transfer(caller, users.Resolve(to), amount)
	if err != nil {
		panic(err)
	}
}

func Approve(spender pusers.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	err := mytoken.Approve(caller, users.Resolve(spender), amount)
	if err != nil {
		panic(err)
	}
}

func TransferFrom(from, to pusers.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	err := mytoken.TransferFrom(caller, users.Resolve(from), users.Resolve(to), amount)
	if err != nil {
		panic(err)
	}
}

func Mint(address pusers.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	err := mytoken.Mint(users.Resolve(address), amount)
	if err != nil {
		panic(err)
	}
}

func Burn(address pusers.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	err := mytoken.Burn(users.Resolve(address), amount)
	if err != nil {
		panic(err)
	}
}

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return mytoken.RenderHome()
	case c == 2 && parts[0] == "balance":
		owner := pusers.AddressOrName(parts[1])
		balance, _ := mytoken.BalanceOf(users.Resolve(owner))
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}

func assertIsAdmin(address std.Address) {
	if address != admin {
		panic("restricted access")
	}
}
```

Detailing what is happening in the above code:
- Calling the `TotalSupply` method would return the total number of tokens minted.
- Calling the `BalanceOf` method would return the total balance of an account.
- Calling the `Allowance` method would set an account as an allowed spender to serve on behalf of the owner.
- Calling the `transfer` method would transfer a configurable amount of token from the calling account to another account, either owned or unowned.
- Calling the `Approve` method would approve a calling account to spend a configurable amount of token on behalf of the token owner.
- Calling the `TransferFrom` method would transfer a configurable amount of token from an account that granted approval to another account, either owned or unowned.
- Calling the `Mint` method would create a configurable number of tokens by the administrator.
- Calling the `Burn` method would destroy a configurable number of tokens by the administrator.
- Calling the `Render` method would return a user's `balance` as a formatted string. Learn more about the `Render`
  method and how it's used [here](../concepts/realms.md).
- Finally, we provide a local function to assert that the calling account is in fact the owner, otherwise panic. This is a very important function that serves to prevent abuse by non-administrators.

## Conclusion

That's it 🎉

You have successfully built a simple GRC20 Realm that is ready to be deployed on the Gno chain and called by users.
In the upcoming guides, we will see how we can develop more complex Realm logic and have them interact with outside tools like a wallet application.
