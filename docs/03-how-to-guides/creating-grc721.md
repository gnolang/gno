---
id: creating-grc721
---

# How to create a GRC721 Token (NFT)

## Overview

This guide shows you how to write a simple **GRC721** Smart Contract, or rather
a [Realm](../concepts/realms.md), in [Gno](../concepts/gno-language.md). 
For actually deploying the Realm, please see the [deployment](deploy.md) guide.

Our **GRC721** Realm will have the following functionality:

- Minting a configurable amount of tokens.
- Keeping track of total token supply.
- Fetching the balance of an account.

## 1. Importing token package

For this realm, we'll want to import the `grc721` package as this will include 
the main functionality of our NFT realm. The package can be found the
`gno.land/p/demo/grc/grc721` path.

[embedmd]:# (../assets/how-to-guides/creating-grc721/mynonfungibletoken-1.gno go)
```go
package mynft

import (
	"std"

	"gno.land/p/demo/grc/grc721"
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
  mynft = grc721.NewBasicNFT("My NFT", "MNFT")

  // Mint 1 million tokens to admin
  mytoken.Mint(admin, 1000000*10000)
}
```

In this code preview, we have:

- Defined and set the value of `mynonfungibletoken` (type `*grc721.basicNFT`) to equal the result of creating a new
  token and configuring its name and symbol.
- Defined and set the value of local variable `admin` to point to a specific gno.land address of type `std.Address`.
- Minted 5 `mynonfungibletoken (MNFT)` and set the administrator as the owner of these tokens

## 2. Adding token functionality

The following section will be about introducing Public functions to expose functionality imported from
the [grc721 package](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo/grc/grc721).

[embedmd]:# (../assets/how-to-guides/creating-grc721/mynonfungibletoken-2.gno go)
```go
func mintNNFT(owner std.Address, n uint64) {
	count := my.TokenCount()
	for i := count; i < count+n; i++ {
		tid := grc721.TokenID(ufmt.Sprintf("%d", i))
		mynonfungibletoken.Mint(owner, tid)
	}
}

// Getters

func BalanceOf(user users.AddressOrName) uint64 {
	balance, err := mynonfungibletoken.BalanceOf(user.Resolve())
	if err != nil {
		panic(err)
	}

	return balance
}

func OwnerOf(tid grc721.TokenID) std.Address {
	owner, err := mynonfungibletoken.OwnerOf(tid)
	if err != nil {
		panic(err)
	}

	return owner
}

func IsApprovedForAll(owner, user users.AddressOrName) bool {
	return mynonfungibletoken.IsApprovedForAll(owner.Resolve(), user.Resolve())
}

func GetApproved(tid grc721.TokenID) std.Address {
	addr, err := mynonfungibletoken.GetApproved(tid)
	if err != nil {
		panic(err)
	}

	return addr
}

// Setters

func Approve(user users.AddressOrName, tid grc721.TokenID) {
	err := mynonfungibletoken.Approve(user.Resolve(), tid)
	if err != nil {
		panic(err)
	}
}

func SetApprovalForAll(user users.AddressOrName, approved bool) {
	err := mynonfungibletoken.SetApprovalForAll(user.Resolve(), approved)
	if err != nil {
		panic(err)
	}
}

func TransferFrom(from, to users.AddressOrName, tid grc721.TokenID) {
	err := mynonfungibletoken.TransferFrom(from.Resolve(), to.Resolve(), tid)
	if err != nil {
		panic(err)
	}
}

// Admin

func Mint(to users.AddressOrName, tid grc721.TokenID) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	err := mynonfungibletoken.Mint(to.Resolve(), tid)
	if err != nil {
		panic(err)
	}
}

func Burn(tid grc721.TokenID) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	err := mynonfungibletoken.Burn(tid)
	if err != nil {
		panic(err)
	}
}

// Render

func Render(path string) string {
	switch {
	case path == "":
		return mynonfungibletoken.RenderHome()
	default:
		return "404\n"
	}
}

// Util

func assertIsAdmin(address std.Address) {
	if address != admin {
		panic("restricted access")
	}
}
```

Detailing what is happening in the above code:

- Calling the **local** `mintNNFT` method would mint a configurable number of tokens to the provided owner's account.
- Calling the `BalanceOf` method would return the total balance of an account.
- Calling the `OwnerOf` method would return the owner of the token based on the ID that is passed into the method.
- Calling the `IsApprovedByAll` method will return true if an operator is approved for all operations by the owner;
  otherwise, returns false.
- Calling the `GetApproved` method will return the address approved to operate on the token.
- Calling the `Approve` method would approve the input address for a particular token.
- Calling the `SetApprovalForAll` method would approve an operating account to operate on all tokens.
- Calling the `TransferFrom` method would transfer a configurable amount of token from an account that granted approval
  to another account, either owned or unowned.
- Calling the `Mint` method would create a configurable number of tokens by the administrator.
- Calling the `Burn` method would destroy a configurable number of tokens by the administrator.
- Calling the `Render` method on success would invoke
  a [`RenderHome`](https://github.com/gnolang/gno/blob/master/examples/gno.land/p/demo/grc/grc721/basic_nft.gno#L353)
  method on the `grc721` instance we instantiated at the top of the file; this method returns a formatted string that
  includes the token: symbol, supply and account balances (`balances avl.Tree`) which is a mapping denoted
  as: `OwnerAddress -> TokenCount`; otherwise returns false and renders a `404`; you can find more information about
  this `Render` method and how it's used [here](../concepts/realms.md).
- Finally, we provide a local function to assert that the calling account is in fact the owner, otherwise panic. This is
  a very important function that serves to prevent abuse by non-administrators.

## Conclusion

That's it 🎉

You have successfully built a simple GRC721 Realm that is ready to be deployed on the Gno chain and called by users.
In the upcoming guides, we will see how we can develop more complex Realm logic and have them interact with outside
tools like a wallet application.
