// This is outdated and is still incomplete. Work in Progress.
// I am considering starting off with an even simpler example,
// of returning funds sent to a test contract.

# What is GNO

## Escrow model

> https://paulx.dev/blog/2021/01/14/programming-on-solana-an-introduction/

```golang
package bank

// Like ERC20 but for Gno.
type Bank20 interface {
	Name() string
	Denom() string
	TotalSupply() int64
	BalanceOf(addr std.Address) int64
	Transfer(to std.Address, value int64) bool
	TransferFrom(from std.Address, to std.Address, value int64) bool
	...
	MakeOrder(from, to std.Address, amount std.Coins) Order
	
}

// NOTE: unexposed struct for security.
type order struct {
	from      Address
	to        Address
	amount    Coins
	processed bool
}

// NOTE: unexposed methods for security.
func (ch *order) string() string {
	return "TODO"
}

// Wraps the internal *order for external use.
type Order struct {
	*order
}

// XXX only exposed for demonstration. TODO unexpose, make full demo.
func NewOrder(from Address, to Address, amount Coins) Order {
	return Order{
		order: &order{
			from:   from,
			to:     to,
			amount: amount,
		},
	}
}

// Panics if error, or already processed.
func (o Order) Execute() {
	if o.order.processed {
		panic("order already processed")
	}
	o.order.processed = true
	// TODO implemement.
}

func (o Order) IsZero() bool {
	return o.order == nil
}

func (o Order) From() Address {
	return o.order.from
}

func (o Order) To() Address {
	return o.order.to
}

func (o Order) Amount() Coins {
	return o.order.amount
}

func (o Order) Processed() bool {
	return o.order.processed
}
```

```golang
package escrow

type EscrowTerms struct {
	PartyA  Address
	PartyB  Address
	AmountA Coins
	AmountB Coins
}

type EscrowContract struct {
	EscrowTerms
	OrderA Order
	OrderB Order
}

func CreateEscrow(terms EscrowTerms) *EscrowContract {
	return &EscrowContract{
		EscrowTerms: terms,
	}
}

func (esc *EscrowContract) SetOrderA(order Order) {
	if !esc.OrderA.IsZero() {
		panic("order-a already set")
	}
	if esc.EscrowTerms.PartyA != order.From() {
		panic("invalid order-a:from mismatch")
	}
	if esc.EscrowTerms.PartyB != order.To() {
		panic("invalid order-a:to mismatch")
	}
	if esc.EscrowTerms.AmountA != order.Amount() {
		panic("invalid order-a amount")
	}
	esc.OrderA = order
}

func (esc *EscrowContract) SetOrderB(order Order) {
	if !esc.OrderB.IsZero() {
		panic("order-b already set")
	}
	if esc.EscrowTerms.PartyB != order.From() {
		panic("invalid order-b:from mismatch")
	}
	if esc.EscrowTerms.PartyA != order.To() {
		panic("invalid order-b:to mismatch")
	}
	if esc.EscrowTerms.AmountB != order.Amount() {
		panic("invalid order-b amount")
	}
	esc.OrderA = order
}

func (esc *EscrowContract) Execute() {
	if esc.OrderA.IsZero() {
		panic("order-a not yet set")
	}
	if esc.OrderB.IsZero() {
		panic("order-b not yet set")
	}
	// NOTE: succeeds atomically.
	esc.OrderA.Execute()
	esc.OrderB.Execute()
}
```

Object oriented programming won market share because it provides encapsulation of logic.
Gno leverages...
We can provide that for smart contracts, 

## No persistence translation.  No databases, only data structures.

How do you persist data to a store?
Answer: you don't.

example with posts...

## Object-level hashing.

## Back to example...

## Where are we?

 * Gnolang
   * Version 0.1: current: persist values and run gno.lang/r/example <-- WE ARE HERE.
   * Version 0.2: 2022: persist types and nodes
 * Tendermint
   * reconcile differences between pkgs/bft and mainline tendermint
   * scenario A: Tendermint/Mainline and Tendermint/GNO, two branches
     scenario B: Tendermint/Mainline depends on Tendermint/Gno (minimal kernel)
 * SDK
   * minimal fork of cosmos-sdk to pkgs/sdk
   * part of Tendermint/GNO.

Philosophy: minimal, secure, (big-O) fast. 

## About the License.

pkgs/bft and pkgs/sdk to become Apache2.0.
Gnolang smart contract logic not yet Apache2.0.
  * Need to balance publishing, adoption, and attribution.
  * Feedback wanted (in github).

## Call to Action.

 * Wanted: volunteers or contractors who can grok Gnolang and contribute.
   - database/language/operating-system/file-system programming required.
   - contact on telegram @cosmosjae.
 * Try to make a test pass from ./tests/challenge/...
 * Document differences between tendermint/gno and mainline tendermint.
 * Create a plan of reconciliation between tendermint/gno and mainline tendermint.
