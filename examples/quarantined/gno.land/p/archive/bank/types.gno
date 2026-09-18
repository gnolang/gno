// TODO: this is an example, and needs to be fixed up and tested.

package bank

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

//----------------------------------------
// Escrow

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
	if !esc.EscrowTerms.AmountA.Equal(order.Amount()) {
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
	if !esc.EscrowTerms.AmountB.Equal(order.Amount()) {
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

//----------------------------------------
// TODO: actually implement these in std package.

type (
	Address string
	Coins   []Coin
	Coin    struct {
		Denom  bool
		Amount int64
	}
)

func (a Coins) Equal(b Coins) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
