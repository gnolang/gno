package orderbook

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestOrderBook_AddOrder(t *testing.T) {
	ob := NewOrderBook()

	order := Order{
		ID:       "1",
		Type:     Buy,
		Quantity: 10,
		Price:    100.0,
	}

	err := ob.AddOrder(order)
	uassert.NoError(t, err, "AddOrder should succeed")

	// Test adding a sell order
	order.Type = Sell
	err = ob.AddOrder(order)
	uassert.NoError(t, err, "AddOrder should succeed for sell order")
}

func TestOrderBook_RemoveOrder(t *testing.T) {
	ob := NewOrderBook()

	order := Order{
		ID:       "1",
		Type:     Buy,
		Quantity: 10,
		Price:    100.0,
	}

	err := ob.AddOrder(order)
	uassert.NoError(t, err, "AddOrder should succeed")

	// Test removing a buy order
	err = ob.RemoveOrder(Buy)
	uassert.NoError(t, err, "RemoveOrder should succeed for buy order")

	// Test removing a sell order
	err = ob.RemoveOrder(Sell)
	uassert.Error(t, err, "RemoveOrder should fail for empty sell orders")
}

func TestOrderBook_PeekOrder(t *testing.T) {
	ob := NewOrderBook()

	order := Order{
		ID:       "1",
		Type:     Buy,
		Quantity: 10,
		Price:    100.0,
	}

	err := ob.AddOrder(order)
	uassert.NoError(t, err, "AddOrder should succeed")

	// Test peeking a buy order
	peekedOrder, err := ob.PeekOrder(Buy)
	uassert.NoError(t, err, "PeekOrder should succeed for buy order")
	uassert.NotEmpty(t, peekedOrder.ID, "Peeked order should not be nil")

	// Test peeking a sell order
	peekedOrder, err = ob.PeekOrder(Sell)
	uassert.Error(t, err, "PeekOrder should fail for empty sell orders")
	uassert.Empty(t, peekedOrder.ID, "Peeked order should be nil for empty sell orders")
}

func TestOrderBook_GetOrderCount(t *testing.T) {
	ob := NewOrderBook()

	order := Order{
		ID:       "1",
		Type:     Buy,
		Quantity: 10,
		Price:    100.0,
	}

	err := ob.AddOrder(order)
	uassert.NoError(t, err, "AddOrder should succeed")

	// Test getting order count for buy orders
	count, err := ob.GetOrderCount(Buy)
	uassert.NoError(t, err, "GetOrderCount should succeed for buy orders")
	uassert.Equal(t, uint64(1), count, "Buy order count should be 1")

	// Test getting order count for sell orders
	count, err = ob.GetOrderCount(Sell)
	uassert.NoError(t, err, "GetOrderCount should succeed for sell orders")
	uassert.Equal(t, uint64(0), count, "Sell order count should be 0")
}

func TestOrderBook_Render(t *testing.T) {
	ob := NewOrderBook()

	order1 := Order{
		ID:       "1",
		Type:     Buy,
		Quantity: 10,
		Price:    100.0,
	}
	order2 := Order{
		ID:       "2",
		Type:     Buy,
		Quantity: 15,
		Price:    95.0,
	}
	order3 := Order{
		ID:       "3",
		Type:     Sell,
		Quantity: 5,
		Price:    110.0,
	}
	order4 := Order{
		ID:       "4",
		Type:     Sell,
		Quantity: 8,
		Price:    115.0,
	}

	// Add orders to the order book
	ob.AddOrder(order1)
	ob.AddOrder(order2)
	ob.AddOrder(order3)
	ob.AddOrder(order4)

	expected := "Buy Orders:\nID: 1, Type: Buy, Quantity: 10, Price: 100.00\nID: 2, Type: Buy, Quantity: 15, Price: 95.00\n\nSell Orders:\nID: 3, Type: Sell, Quantity: 5, Price: 110.00\nID: 4, Type: Sell, Quantity: 8, Price: 115.00\n"

	result := ob.Render(3) // Limit to 3 orders per type

	uassert.Equal(t, expected, result, "Render() result does not match expected")
}
