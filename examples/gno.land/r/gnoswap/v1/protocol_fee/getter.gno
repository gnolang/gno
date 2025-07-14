package protocol_fee

import "gno.land/p/demo/avl"

// TokenList returns only the list of token path.
// If positive is true, it returns only the token path with amount > 0.
// If positive is false, it returns all the token path.
func TokenList(positive bool) []string {
	tokens := []string{}
	for tokenPath, amount := range protocolFeeState.TokenListWithAmount {
		if positive && amount == 0 {
			continue
		}
		tokens = append(tokens, tokenPath)
	}
	return tokens
}

// GetTokenListWithAmount returns the token path and amount.
func GetTokenListWithAmount() map[string]uint64 {
	return protocolFeeState.TokenListWithAmount
}

// GetAmountOfToken returns the amount of token.
func GetAmountOfToken(tokenPath string) uint64 {
	amount, exists := protocolFeeState.TokenListWithAmount[tokenPath]
	if !exists {
		return 0
	}
	return amount
}

// GetAccuTransferToGovStaker returns the accuToGovStaker.
func GetAccuTransferToGovStaker() *avl.Tree {
	return protocolFeeState.AccuToGovStaker
}

// GetAccuTransferToDevOps returns the accuToDevOps.
func GetAccuTransferToDevOps() *avl.Tree {
	return protocolFeeState.AccuToDevOps
}

// GetAccuTransferToGovStakerByTokenPath returns the accumulated transfer to gov/staker by token path.
func GetAccuTransferToGovStakerByTokenPath(path string) uint64 {
	return protocolFeeState.GetAccuTransferToGovStakerByTokenPath(path)
}

// GetAccuTransferToDevOpsByTokenPath returns the accumulated transfer to devOps by token path.
func GetAccuTransferToDevOpsByTokenPath(path string) uint64 {
	return protocolFeeState.GetAccuTransferToDevOpsByTokenPath(path)
}

// GetHistoryOfDistributedToGovStakerByTokenPath returns the history of distributed to gov/staker by token path.
func GetHistoryOfDistributedToGovStakerByTokenPath(path string) uint64 {
	amountI, exists := protocolFeeState.DistributedToGovStakerHistory.Get(path)
	if !exists {
		return 0
	}
	return amountI.(uint64)
}

// GetHistoryOfDistributedToDevOpsByTokenPath returns the history of distributed to devOps by token path.
func GetHistoryOfDistributedToDevOpsByTokenPath(path string) uint64 {
	amountI, exists := protocolFeeState.DistributedToDevOpsHistory.Get(path)
	if !exists {
		return 0
	}
	return amountI.(uint64)
}
