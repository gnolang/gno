package treasury

import (
	"errors"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/list"
	"gno.land/p/nt/mux/v0"
)

// Treasury is the main structure that holds all bankers and their payment
// history. It also provides a router for rendering the treasury pages.
type Treasury struct {
	bankers *avl.Tree // string -> *bankerRecord
	router  *mux.Router
}

// bankerRecord holds a Banker and its payment history.
type bankerRecord struct {
	banker  Banker
	history list.List // List of Payment.
}

// Banker is an interface that allows for banking operations.
type Banker interface {
	ID() string          // Get the ID of the banker.
	Send(Payment) error  // Send a payment to a recipient.
	Balances() []Balance // Get the balances of the banker.
	Address() string     // Get the address of the banker to receive payments.
}

// Payment is an interface that allows getting details about a payment.
type Payment interface {
	BankerID() string // Get the ID of the banker that can process this payment.
	String() string   // Get a string representation of the payment.
}

// Balance represents the balance of an asset held by a Banker.
type Balance struct {
	Denom  string // The denomination of the asset
	Amount int64  // The amount of the asset
}

// Common Banker errors.
var (
	ErrCurrentRealmIsNotOwner = errors.New("current realm is not the owner of the banker")
	ErrNoOwnerProvided        = errors.New("no owner provided")
	ErrInvalidPaymentType     = errors.New("invalid payment type")
)
