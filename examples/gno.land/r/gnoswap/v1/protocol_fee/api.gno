package protocol_fee

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/json"

	"gno.land/r/gnoswap/v1/common"
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
	historyTransferList := distributedToGovStakerHistory()
	if historyTransferList == nil {
		return ""
	}

	return marshal(buildByAvlTree(historyTransferList))
}

func ApiGetHistoryTransferToDevOps() string {
	historyTransferList := distributedToDevOpsHistory()
	if historyTransferList == nil {
		return ""
	}
	return marshal(buildByAvlTree(historyTransferList))
}

func buildByAvlTree(tree *avl.Tree) *json.Node {
	data := json.Builder().
		WriteString("height", formatInt(std.ChainHeight())).
		WriteString("now", formatInt(time.Now().Unix()))

	tree.Iterate("", "", func(key string, value any) bool {
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

// ApiGetActualBalance returns all tokens with their current balances (recorded + unrecorded)
func ApiGetActualBalance() string {
	tokenMap := make(map[string]uint64)

	// get all recorded tokens
	for token := range protocolFeeState.TokenListWithAmount {
		actualBalance := uint64(common.BalanceOf(token, protocolFeeAddr))
		tokenMap[token] = actualBalance
	}

	// only include tokens that are already recorded.
	// check for any tokens that have balance will requires
	// iterating through known tokens or having a registry.
	data := json.Builder().
		WriteString("height", formatInt(std.ChainHeight())).
		WriteString("now", formatInt(time.Now().Unix()))

	for token, balance := range tokenMap {
		data.WriteString(token, formatUint(balance))
	}

	return marshal(data.Node())
}

// ApiGetRecordedBalance returns the recorded tokens and their amounts
func ApiGetRecordedBalance() string {
	tokenList := GetTokenListWithAmount()
	if tokenList == nil {
		return ""
	}

	data := json.Builder().
		WriteString("height", formatInt(std.ChainHeight())).
		WriteString("now", formatInt(time.Now().Unix()))

	for token, amount := range tokenList {
		data.WriteString(token, formatUint(amount))
	}

	return marshal(data.Node())
}

// ApiGetUnrecordedBalance returns tokens with unrecorded balances
func ApiGetUnrecordedBalance() string {
	unrecordedMap := make(map[string]uint64)

	// Check all recorded tokens for discrepancies
	for token, recordedAmount := range protocolFeeState.TokenListWithAmount {
		actualBalance := uint64(common.BalanceOf(token, protocolFeeAddr))
		if actualBalance > recordedAmount {
			unrecordedAmount := actualBalance - recordedAmount
			unrecordedMap[token] = unrecordedAmount
		}
	}

	data := json.Builder().
		WriteString("height", formatInt(std.ChainHeight())).
		WriteString("now", formatInt(time.Now().Unix()))

	for token, unrecordedBalance := range unrecordedMap {
		data.WriteString(token, formatUint(unrecordedBalance))
	}

	return marshal(data.Node())
}
