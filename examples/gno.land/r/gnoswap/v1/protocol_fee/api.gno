package protocol_fee

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/json"
)

func ApiGetTokenListWithAmountOverZero() string {
	tokenList := GetTokenListWithAmount()
	if tokenList == nil {
		return ""
	}

	data := json.Builder().
		WriteString("height", formatInt(std.ChainHeight())).
		WriteString("now", formatInt(time.Now().Unix()))

	for token, amount := range tokenList {
		if amount == 0 {
			continue
		}
		data.WriteString(token, formatUint(amount))
	}

	return marshal(data.Node())
}

func ApiGetAccuTransferToGovStaker() string {
	distributedAmountList := GetAccuTransferToGovStaker()
	if distributedAmountList == nil {
		return ""
	}

	return marshal(buildByAvlTree(distributedAmountList))
}

func ApiGetAccuTransferToDevOps() string {
	distributedAmountList := GetAccuTransferToDevOps()
	if distributedAmountList == nil {
		return ""
	}

	return marshal(buildByAvlTree(distributedAmountList))
}

func ApiGetHistoryTransferToGovStaker() string {
	historyTransferList := GetHistoryOfDistributedToGovStaker()
	if historyTransferList == nil {
		return ""
	}

	return marshal(buildByAvlTree(historyTransferList))
}

func ApiGetHistoryTransferToDevOps() string {
	historyTransferList := GetHistoryOfDistributedToDevOps()
	if historyTransferList == nil {
		return ""
	}
	return marshal(buildByAvlTree(historyTransferList))
}

func buildByAvlTree(tree *avl.Tree) *json.Node {
	data := json.Builder().
		WriteString("height", formatInt(std.ChainHeight())).
		WriteString("now", formatInt(time.Now().Unix()))

	tree.Iterate("", "", func(key string, value interface{}) bool {
		data.WriteString(key, formatUint(value.(uint64)))
		return false
	})

	return data.Node()
}

func formatUint(v uint64) string {
	return strconv.FormatUint(v, 10)
}

func formatInt(v int64) string {
	return strconv.FormatInt(v, 10)
}

func marshal(node *json.Node) string {
	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}
