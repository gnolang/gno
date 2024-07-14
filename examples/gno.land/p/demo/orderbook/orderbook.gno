package orderbook

import (
	"fmt"

	"gno.land/p/demo/dequeue"
	"gno.land/p/demo/ufmt"
)

type OrderType int

const (
	Buy OrderType = iota
	Sell
)

type Order struct {
	ID       string
	Type     OrderType
	Quantity int
	Price    float64
}

type OrderBook struct {
	buyOrders  *dequeue.Dequeue
	sellOrders *dequeue.Dequeue
}

// NewOrderBook initializes a new order book
func NewOrderBook() *OrderBook {
	return &OrderBook{
		buyOrders:  dequeue.New(),
		sellOrders: dequeue.New(),
	}
}

// AddOrder adds a new order to the order book
func (ob *OrderBook) AddOrder(order Order) error {
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
func (ob *OrderBook) RemoveOrder(orderType OrderType) error {
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
func (ob *OrderBook) PeekOrder(orderType OrderType) (Order, error) {
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
func (ob *OrderBook) GetOrderCount(orderType OrderType) (uint64, error) {
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
func (ob *OrderBook) Render(n int) string {
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
		output += ufmt.Sprintf("ID: %s, Type: Buy, Quantity: %d, Price: %.2f\n", order.ID, order.Quantity, order.Price)
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
		output += ufmt.Sprintf("ID: %s, Type: Sell, Quantity: %d, Price: %.2f\n", order.ID, order.Quantity, order.Price)
	}

	return output
}
