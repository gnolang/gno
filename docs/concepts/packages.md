---
id: packages
---

# Packages

Packages aim to encompass functionalities that are more closely aligned with the characteristics and capabilities of realms, as opposed to standard libraries. As opposed to realms, they are stateless.

The full list of pre-deployed available packages can be found under the [demo package](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo). Below are some of the most commonly used packages.

## `avl`

In Go, the classic key/value data type is represented by the `map` construct. However, while Gno also supports the use of `map`, it is not a viable option as it lacks determinism due to its non-sequential nature.
 
To address this issue, Gno implements the [AVL Tree](https://en.wikipedia.org/wiki/AVL\_tree) (Adelson-Velsky-Landis Tree) as a solution. The AVL Tree is a self-balancing binary search tree.

The `avl` package comprises a set of functions that can manipulate the leaves and nodes of the AVL Tree.

## `grc20`

Gno includes an implementation of the `erc20` fungible token standard referred to as `grc20`. The interfaces of `grc20` are as follows:

[embedmd]:# (../assets/explanation/packages/pkg-1.gno go)
```go
func TotalSupply() uint64
func BalanceOf(account std.Address) uint64
func Transfer(to std.Address, amount uint64)
func Approve(spender std.Address, amount uint64)
func TransferFrom(from, to std.Address, amount uint64)
func Allowance(owner, spender std.Address) uint64
```

The role of each function is as follows:

* `TotalSupply`: Returns the total supply of the token.
* `BalanceOf`: Returns the balance of tokens of an account.
* `Transfer`: Transfers specific `amount` of tokens from the `caller` of the function to the `to` address.
* `Approve`: Grants the `spender`(also referred to as `operator`) with the ability to send specific `amount` of the `caller`'s (also referred to as `owner`) tokens on behalf of the `caller`.
* `TransferFrom`: Can be called by the `operator` to send specific `amount` of `owner`'s tokens from the `owner`'s address to the `to` address.
* `Allowance`: Returns the number of tokens approved to the `spender` by the `owner`.

Two types of contracts exist in`grc20`:

1. `AdminToken`
   - Implements the token factory with `Helper` functions.
   - The underlying struct should not be exposed to users. However, it can be typecasted as UserToken using the `GRC20()` method.
2. `UserToken`
   - Implements the `IGRC20` interface.
   - The underlying struct can be exposed to users. Created with the `GRC20()` method of `adminToken`.

## `grc721`

Gno includes an implementation of the `erc721` non-fungible token standard referred to as `grc721`. The interfaces of `grc721` are as follows:

[embedmd]:# (../assets/explanation/packages/pkg-2.gno go)
```go
// functions that work similarly to those of grc20
func BalanceOf(owner std.Address) (uint64, error)
func Approve(approved std.Address, tid TokenID) error
func TransferFrom(from, to std.Address, tid TokenID) error

// functions unique to grc721
func OwnerOf(tid TokenID) (std.Address, error)
func SafeTransferFrom(from, to std.Address, tid TokenID) error
func SetApprovalForAll(operator std.Address, approved bool) error
func GetApproved(tid TokenID) (std.Address, error)
func IsApprovedForAll(owner, operator std.Address) bool
```

`grc721` contains a new set of functions:

* `OwnerOf`: Returns the `owner`'s address of a token specified by its `TokenID`.
* `SafeTransferFrom`: Equivalent to the `TransferFrom` function of `grc20`.
  * The `Safe` prefix indicates that the function runs a check to ensure that the `to` address is a valid address that can receive tokens.
  * As you can see from the [code](https://github.com/gnolang/gno/blob/master/examples/gno.land/p/demo/grc/grc721/basic\_nft.gno#L341), the concept of `Safe` has yet to be implemented.
* `SetApprovalForAll`: Approves all tokens owned by the `owner` to an `operator`.
  * You may not set multiple `operator`s.
* `GetApproved`: Returns the `address` of the `operator` for a token, specified with its `ID`.
* `IsApprovedForAll`: Returns if all NFTs of the `owner` have been approved to the `operator`.

## `testutils`

The `testutils` package contains a set of functions that comes in handy when testing realms. The sample function below is the commonly used `TestAddress` function that generates a random address.

[embedmd]:# (../assets/explanation/packages/pkg-3.gno go)
```go
func TestAddress(name string) std.Address {
	if len(name) > std.RawAddressSize {
		panic("address name cannot be greater than std.AddressSize bytes")
	}
	addr := std.RawAddress{}
	// TODO: use strings.RepeatString or similar.
	// NOTE: I miss python's "".Join().
	blanks := "____________________"
	copy(addr[:], []byte(blanks))
	copy(addr[:], []byte(name))
	return std.Address(std.EncodeBech32("g", addr))
}
```

The code takes the `name` as the input and creates a random address. Below is a list of examples where it's used in the test case of the `foo20` realm.

[embedmd]:# (../assets/explanation/packages/pkg-4.gno go)
```go
admin := users.AddressOrName("g1tntwtvzrkt2gex69f0pttan0fp05zmeg5yykv8")
test2 := users.AddressOrName(testutils.TestAddress("test2"))
recv := users.AddressOrName(testutils.TestAddress("recv"))
normal := users.AddressOrName(testutils.TestAddress("normal"))
owner := users.AddressOrName(testutils.TestAddress("owner"))
spender := users.AddressOrName(testutils.TestAddress("spender"))
recv2 := users.AddressOrName(testutils.TestAddress("recv2"))
mibu := users.AddressOrName(testutils.TestAddress("mint_burn"))
```
