---
id: creating-grc20
---

# How to Create a GRC20 Token
## Overview

This guide shows you how to write a simple **GRC20**
a [realm](../concepts/realms.md) in [Gno](../concepts/gno-language.md). For 
actually deploying the Realm, please see the [deployment](deploy.md) guide.

Our **GRC20** realm will follow the GRC20 standard, which is an adaptation
of Ethereum's ERC20 fungible token standard.

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
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
)

var (
	mytoken *grc20.Token     // main token object
	o       *ownable.Ownable // ownable to store admin of token
)

func init() {
	// Instantiate ownable to handle administration
	o = ownable.New()

	// Instantiate token
	mytoken = grc20.NewGRC20Token("My Token", "MYT", 4)

	// Mint 1M tokens to admin
	mytoken.Mint(o.Owner(), uint64(1_000_000*10000))
}
```

The code snippet above does the following:
- Defines a new token variable, `mytoken` of type `*grc20.Token`
- Defines a new ownable variable of type `*ownable.Ownable` for handling
administration of the realm, which sets the admin to the deployer of the realm
automatically
- Initializes `mytoken` as a new GRC20 token, and set its name, symbol, and 
decimal values,
- Mint 1 million units of `My Token` and assign them to the admin's address.

## 2. Adding token functionality

In order to call exported functions from the `grc20` package, we also need to 
expose them in the realm. Let's go through all functions in the GRC20 package,
one by one:

```go
// Name returns the name of the token
func Name() string {
	return mytoken.Name()
}
```

Calling the `Name` method would return the name of the token.

```go
// Symbol returns the name of the token
func Symbol() string {
	return mytoken.Symbol()
}
```

Calling the `Symbol` method would return the symbol of the token.

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
    return mytoken.BalanceOf(account)
}
```

Calling the `BalanceOf` method would return the total balance of an account.

```go
// Allowance returns the allowance of spender on owner's balance
func Allowance(owner, spender std.Address) uint64 {
    return mytoken.Allowance(owner, spender)
}
```

Calling the `Allowance` method will return the amount `spender` is allowed to spend
from `owner`'s balance.

```go
// Transfer transfers amount from caller to recipient
func Transfer(recipient std.Address, amount uint64) {
    mytoken.Transfer(recipient, amount)
}
```

Calling the `Transfer` method transfers amount of token from the calling account 
to the recipient account.

```go
// Approve approves amount of caller's tokens to be spent by spender
func Approve(spender std.Address, amount uint64) {
    mytoken.Approve(spender, amount)
}
```

Calling the `Approve` method approves `spender` to spend `amount` from the caller's
balance of tokens.

```go
// TransferFrom transfers `value` amount of tokens from address `from` to address `to`, and MUST fire the Transfer event
func TransferFrom(from, to std.Address, amount uint64) {
    mytoken.TransferFrom(from, to, amount)
}
```

Calling the `TransferFrom` method moves `amount` of tokens from `sender` to 
`recipient` using the allowance mechanism. `amount` is then deducted from the
caller’s allowance.

```go
// Mint mints `amount` of tokens to `address`, only callable by owner
func Mint(address std.Address, amount uint64) {
    if err := o.CallerIsOwner(); err != nil {
        panic("only owner can mint tokens")
    }

    mytoken.Mint(address, amount)
}

```

Calling the `Mint` method checks if the caller is the admin of the token 
creates `amount` of tokens and assigns them to `address`, increasing the total supply.

```go
// Burn burns `amount` of tokens from `address`, only callable by owner
func Burn(address std.Address, amount uint64) {
    if err := o.CallerIsOwner(); err != nil {
        panic("only owner can burn tokens")
    }
    
    mytoken.Burn(address, amount)
}
```

Calling the `Burn` method checks if the caller is the admin of the token and
burns `amount` of tokens from the balance of `address`, decreasing the total supply.

```go
func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return mytoken.RenderHome()
	case c == 2 && parts[0] == "balance": // pkgpath:balance/address
		owner := std.Address(parts[1])
		return ufmt.Sprintf("%d\n", mytoken.BalanceOf(owner))
	default:
		return "404\n"
	}
}
```

Calling the `Render` method returns a general render of the token realm, or
if given a specific address, the user's `balance` as a formatted string.

You can view the full code on [this Playground link](https://play.gno.land/p/kUOn0OG7oLL).
If you want to deploy it, do so on the [Portal Loop](../concepts/portal-loop.md).

## Conclusion

That's it 🎉

You have successfully built a simple GRC20 realm that is ready to be deployed on 
the Gno chain and called by users.
In the upcoming guides, we will see how we can develop more complex realm logic
and have them interact with outside tools like a wallet application.
