package minidex

import (
	"std"

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
		ordersTree = &avl.Tree{} // index -> order
	}
	ordersTree.(*avl.Tree).Set(orderIndex.Binary(), order)
	dex.orders.Set(pair, ordersTree)

	std.Emit(
		"order_placed",
		"trader", trader.String(),
		"tokenFrom", tokenFrom.GetName(), // XXX: fqname
		"tokenTo", tokenTo.GetName(),
		"amount", ufmt.Sprintf("%d", amount),
		"price", ufmt.Sprintf("%d", price),
	)
}

// MatchPairOrders matches orders for a given pair.
func (dex *DEX) MatchPairOrders(pair string) int {
	ordersTree, ok := dex.orders.Get(pair)
	if !ok {
		return 0
	}

	matched := 0
	orders := ordersTree.(*avl.Tree)
	orders.Iterate("", "", func(key1 string, value interface{}) bool {
		order1 := value.(*Order)
		orders.Iterate("", "", func(key2 string, value2 interface{}) bool {
			order2 := value2.(*Order)
			if order1.tokenFrom == order2.tokenTo && order1.tokenTo == order2.tokenFrom {
				amount1 := min(order1.amount, order2.amount)
				amount2 := amount1 * order1.price / order2.price
				order1.amount -= amount1
				order2.amount -= amount2
				banker1 := grc20.AccountBanker(order1.tokenFrom, "")
				banker2 := grc20.AccountBanker(order2.tokenFrom, "")
				banker1.Transfer(order2.trader, amount1)
				banker2.Transfer(order1.trader, amount2)
				matched++
				std.Emit(
					"trade_executed",
					"trader1", order1.trader.String(),
					"trader2", order2.trader.String(),
					"token1", order1.tokenFrom.GetName(),
					"token2", order2.tokenFrom.GetName(),
					"amount1", ufmt.Sprintf("%d", amount1),
					"amount2", ufmt.Sprintf("%d", amount2),
				)
				if order1.amount == 0 {
					orders.Remove(key1)
				}
				if order2.amount == 0 {
					orders.Remove(key2)
				}
				return false
			}
			return true
		})
		return true
	})
	return matched
}

// CancelOrders cancels all orders for the caller and refunds the tokens.
func (dex *DEX) CancelOrders() {
	trader := std.PrevRealm().Addr()
	dex.orders.Iterate("", "", func(pair string, ordersTreeI interface{}) bool {
		ordersTree := ordersTreeI.(*avl.Tree)
		ordersTree.Iterate("", "", func(key string, value interface{}) bool {
			order := value.(*Order)
			if order.trader == trader {
				userBanker := grc20.AccountBanker(order.tokenFrom, "")
				err := userBanker.Transfer(trader, order.amount)
				checkErr(err, "cannot refund tokens")
				ordersTree.Remove(key)
			}
			return true
		})
		return true
	})
	std.Emit(
		"orders_cancelled",
		"trader", trader.String(),
	)
}

// AdminClearOrders clears all orders in the DEX (admin function).
func (dex *DEX) AdminClearOrders() {
	dex.orders = avl.Tree{}
	std.Emit("orders_cleared")
}

func (dex DEX) Size() int {
	size := 0
	dex.orders.Iterate("", "", func(pair string, ordersTreeI interface{}) bool {
		ordersTree := ordersTreeI.(*avl.Tree)
		size += ordersTree.Size()
		return false
	})
	return size
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
