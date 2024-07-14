package orderbook

import (
	"gno.land/p/demo/dequeue"
	"gno.land/p/demo/ufmt"
)

type OrderType int

const (
	Buy OrderType = iota
	Sell
)

// Order represents a single order with ID, type (Buy or Sell), quantity, and price
type Order struct {
	ID       string    // ID of the order
	Type     OrderType // Type of the order (Buy or Sell)
	Quantity int       // Quantity of the order
	Price    uint64    // Price of the order (temporary uint64 due to lack of float support)
}

// OrderBook defines the interface for an order book
type OrderBook interface {
	AddOrder(order Order) error
	RemoveOrder(orderType OrderType) error
	PeekOrder(orderType OrderType) (Order, error)
	GetOrderCount(orderType OrderType) (uint64, error)
	Render(n int) string
}

// sequentialOrderBook represents an order book prioritized by order type (Buy or Sell)
type sequentialOrderBook struct {
	buyOrders  *dequeue.Dequeue
	sellOrders *dequeue.Dequeue
}

// pricePriorityOrderBook represents an order book prioritized by price
type pricePriorityOrderBook struct {
	buyOrders  *dequeue.Dequeue
	sellOrders *dequeue.Dequeue
}

// quantityOrderBook represents an order book prioritized by quantity
type quantityPriorityOrderBook struct {
	buyOrders  *dequeue.Dequeue
	sellOrders *dequeue.Dequeue
}

// NewOrderBook initializes a new priority order book
func NewSequentialOrderBook() OrderBook {
	return &sequentialOrderBook{
		buyOrders:  dequeue.New(),
		sellOrders: dequeue.New(),
	}
}

var _ OrderBook = (*sequentialOrderBook)(nil)

// AddOrder adds a new order to the order book
func (ob *sequentialOrderBook) AddOrder(order Order) error {
	switch order.Type {
	case Buy:
		err := ob.buyOrders.PushBack(order)
		if err != nil {
			return ErrFailedToAddOrder
		}
	case Sell:
		err := ob.sellOrders.PushBack(order)
		if err != nil {
			return ErrFailedToAddOrder
		}
	default:
		return ErrUnknownOrderType
	}
	return nil
}

// RemoveOrder removes the first order from the order book
func (ob *sequentialOrderBook) RemoveOrder(orderType OrderType) error {
	var err error

	switch orderType {
	case Buy:
		_, err = ob.buyOrders.PopFront()
		if err != nil {
			return ErrFailedToRemoveOrder
		}
	case Sell:
		_, err = ob.sellOrders.PopFront()
		if err != nil {
			return ErrFailedToRemoveOrder
		}
	default:
		return ErrUnknownOrderType
	}

	return nil
}

// PeekOrder returns the first order in the order book without removing it
func (ob *sequentialOrderBook) PeekOrder(orderType OrderType) (Order, error) {
	switch orderType {
	case Buy:
		order, err := ob.buyOrders.Front()
		if err != nil {
			return Order{}, ErrFailedToPeekOrder
		}
		return order.(Order), nil
	case Sell:
		order, err := ob.sellOrders.Front()
		if err != nil {
			return Order{}, ErrFailedToPeekOrder
		}
		return order.(Order), nil
	default:
		return Order{}, ErrUnknownOrderType
	}
}

// GetOrderCount returns the number of orders in the order book
func (ob *sequentialOrderBook) GetOrderCount(orderType OrderType) (uint64, error) {
	switch orderType {
	case Buy:
		return ob.buyOrders.Length(), nil
	case Sell:
		return ob.sellOrders.Length(), nil
	default:
		return 0, ErrUnknownOrderType
	}
}

// Render returns a string representation of the order book with up to n orders per type
func (ob *sequentialOrderBook) Render(n int) string {
	var output string

	// buy orders
	output += "Buy Orders:\n"
	buyLength := ob.buyOrders.Length()
	for i := uint64(0); i < buyLength && i < uint64(n); i++ {
		value, err := ob.buyOrders.At(i)
		if err != nil {
			continue
		}
		order := value.(Order)
		output += ufmt.Sprintf("ID: %s, Quantity: %d, Price: %d\n", order.ID, order.Quantity, order.Price)
	}
	output += "\n"

	// sell orders
	output += "Sell Orders:\n"
	sellLength := ob.sellOrders.Length()
	for i := uint64(0); i < sellLength && i < uint64(n); i++ {
		value, err := ob.sellOrders.At(i)
		if err != nil {
			continue
		}
		order := value.(Order)
		output += ufmt.Sprintf("ID: %s, Quantity: %d, Price: %d\n", order.ID, order.Quantity, order.Price)
	}

	return output
}
