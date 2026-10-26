# Coins and tokens

gno.land has two kinds of tokens: Coins, which the chain keeps, and GRC20
tokens, which are contracts.

## Choosing between Coins and GRC20 tokens

Pick Coins when a token only needs to be held and sent. The chain moves them
with no contract code, and a user pays a realm by attaching them with `-send`
to a function that reads `unsafe.OriginSend()`.

Pick GRC20 when a token needs rules of its own: an allowance a spender draws
on, a fee or a cap on transfers, or balances a DAO counts as votes. Those rules
live in the token's contract.

To get both, wrap a Coin in a GRC20 token, as described
[below](#wrapping-coins).

## Coins

Coins are balances the chain keeps outside the GnoVM. Sending them or reading a
balance runs no contract code. A realm can issue its own denom through a
[banker](./gno-stdlibs.md#banker), and can [burn](./gno-stdlibs.md#removecoin)
that denom from any holder. A denom listed in the chain's
[`restricted_denoms`](../../examples/gno.land/r/sys/params/unlock.gno)
parameter can be sent only from whitelisted accounts.

## Verifying inbound Coin payments

A user pays your realm by attaching coins to a `maketx call` with `-send`.
In the called function, first check that the user called you directly with
`cur.Previous().IsUserCall()`, then compare the amount sent, read with
[`unsafe.OriginSend()`](./gno-stdlibs.md#originsend), to your price:

```go
import "chain/runtime/unsafe"

const price = 1_000_000 // ugnot

func Buy(cur realm) {
	if !cur.Previous().IsUserCall() {
		panic("must be called directly by a user")
	}
	if unsafe.OriginSend().AmountOf("ugnot") != price {
		panic("wrong payment amount")
	}
	// ...
}
```

The [security guide](./gno-security-guide.md#55-isuser-for-payment-guards)
explains why the first check matters. Without the second, a call that attaches
no coins succeeds.

## GRC20 tokens

GRC20 is Gno's version of ERC20, with `Approve` and `TransferFrom`. Every
transfer runs the token's contract, which keeps control of the rules. That lets
a token gate access, sit in a vault, or count votes in a DAO.

`Approve` replaces the allowance rather than adding to it. To change one that
isn't zero, set it to 0 first. Otherwise the spender can spend both the old and
the new amount.

```go
import "gno.land/p/nt/grc20/v0"

var (
	Token         *grc20.Token
	privateLedger *grc20.PrivateLedger
	// Keep the teller unexported: it is a spend capability. It is built from
	// the ledger, which never leaves this realm, and it only works here.
	userTeller grc20.Teller
)

func init(cur realm) {
	Token, privateLedger = grc20.NewToken("Foo Token", "FOO", 4, 0, cur)
	userTeller = privateLedger.CallerTeller()
}

// Transfer moves the caller's own tokens (userTeller debits cur.Previous()).
func Transfer(cur realm, to address, amount int64) {
	if err := userTeller.Transfer(0, cur, to, amount); err != nil {
		panic(err)
	}
}
```

See also [`gno.land/r/demo/defi/foo20`](../../examples/gno.land/r/demo/defi/foo20).

## Wrapping Coins

A realm can wrap a coin in a GRC20 token. A user deposits the coin and gets the
token, which then follows the contract's rules. Withdrawing burns the token and
returns the coin.
[`gno.land/r/gnoland/wugnot`](../../examples/gno.land/r/gnoland/wugnot) wraps
`ugnot` this way.
