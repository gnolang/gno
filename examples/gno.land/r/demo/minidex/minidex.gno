package minidex

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

// Order represents an order in the DEX.
type Order struct {
	trader    std.Address
	tokenFrom *grc20.Bank
	tokenTo   *grc20.Bank
	amount    uint64
	price     uint64
}

// DEX represents the decentralized exchange.
type DEX struct {
	index  seqid.ID
	orders avl.Tree // pair -> orders
}

// NewDEX creates a new DEX instance.
func NewDEX() *DEX {
	return &DEX{
		orders: avl.Tree{},
	}
}

// PlaceOrder places a new order and matches orders.
func (dex *DEX) PlaceOrder(tokenFrom, tokenTo *grc20.Bank, amount, price uint64) {
	trader := std.PrevRealm().Addr()
	curAddr := std.CurrentRealm().Addr()
	userBanker := grc20.AccountBanker(tokenFrom, "")

	allowance := userBanker.Allowance(trader, curAddr)
	require(allowance >= amount, "insufficient allowance")
	err := userBanker.TransferFrom(trader, curAddr, amount)
	checkErr(err, "cannot retrieve tokens from allowance")

	order := &Order{
		trader:    trader,
		tokenFrom: tokenFrom,
		tokenTo:   tokenTo,
		amount:    amount,
		price:     price,
	}
	pair := tokenFrom.GetName() + "/" + tokenTo.GetName()
	orderIndex := dex.index.Next()

	ordersTree, ok := dex.orders.Get(pair)
	if !ok {
		ordersTree = avl.Tree{} // index -> order
	}
	ordersTree.(*avl.Tree).Set(orderIndex.Binary(), order)
	dex.orders.Set(pair, ordersTree)

	std.Emit(
		"order_placed",
		"trader", trader.String(),
		"tokenFrom", tokenFrom.GetName(), // XXX: fqname
		"tokenTo", tokenTo.GetName(),
		"amount", strconv.Itoa(amount),
		"price", strconv.Itoa(price),
	)

	dex.MatchOrders(pair)
}

// MatchOrders matches orders for a given pair.
func (dex *DEX) MatchOrders(pair string) {
	ordersTree, ok := dex.orders.Get(pair)
	if !ok {
		return
	}

	orders := ordersTree.(*avl.Tree)
	orders.Iterate("", "", func(key string, value interface{}) bool {
		order1 := value.(*Order)
		orders.Iterate("", "", func(key2 string, value2 interface{}) bool {
			order2 := value2.(*Order)
			if order1.tokenFrom == order2.tokenTo && order1.tokenTo == order2.tokenFrom {
				tradeAmount := min(order1.amount, order2.amount)
				order1.amount -= tradeAmount
				order2.amount -= tradeAmount
				order1.tokenFrom.TransferFrom(order2.trader, order1.trader, tradeAmount)
				order2.tokenFrom.TransferFrom(order1.trader, order2.trader, tradeAmount*order1.price/order2.price)
				std.Emit(
					"trade_executed",
					"trader1", order1.trader,
					"trader2", order2.trader,
					"tokenFrom1", order1.tokenFrom.GetName(),
					"tokenTo1", order1.tokenTo.GetName(),
					"amount", tradeAmount,
					"price", order1.price,
				)
				if order1.amount == 0 {
					orders.Delete(key)
				}
				if order2.amount == 0 {
					orders.Delete(key2)
				}
				return false
			}
			return true
		})
		return true
	})
}

// CancelOrders cancels all orders for the caller and refunds the tokens.
func (dex *DEX) CancelOrders() {
	trader := std.GetOrigCaller()
	dex.orders.Iterate("", "", func(pair string, ordersTreeI interface{}) bool {
		ordersTree := ordersTreeI.(*avl.Tree)
		ordersTree.Iterate("", "", func(key string, value interface{}) bool {
			order := value.(*Order)
			if order.trader == trader {
				userBanker := grc20.AccountBanker(order.tokenFrom.(*grc20.Bank), "")
				err := userBanker.Transfer(trader, order.amount)
				checkErr(err, "cannot refund tokens")
				ordersTree.Delete(key)
			}
			return true
		})
		return true
	})
	std.Emit("orders_cancelled", "trader", trader)
}

// AdminClearOrders clears all orders in the DEX (admin function).
func (dex *DEX) AdminClearOrders() {
	dex.orders = avl.NewTree()
	std.Emit("orders_cleared")
}

// Render returns the current state of the DEX.
func (dex *DEX) Render(_ string) string {
	s := "Orders:\n"
	dex.orders.Iterate("", "", func(pair string, ordersTreeI interface{}) bool {
		ordersTree := ordersTreeI.(*avl.Tree)
		ordersTree.Iterate("", "", func(key string, value interface{}) bool {
			order := value.(*Order)
			s += ufmt.Sprintf("%s -> %s %d at %d\n", order.tokenFrom.GetName(), order.tokenTo.GetName(), order.amount, order.price)
			return true
		})
		return true
	})
	return s
}

// min returns the smaller of two uint64 values.
func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// require checks a condition and panics with a message if the condition is false.
func require(check bool, msg string) {
	if !check {
		panic(msg)
	}
}

// checkErr checks for an error and panics with a message if the error is not nil.
func checkErr(err error, msg string) {
	if err != nil {
		panic(msg + " - " + err.Error())
	}
}
