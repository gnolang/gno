package protocol_fee

import (
	"gno.land/p/demo/avl"
)

// TokenList returns only the list of token path.
// If positive is true, it returns only the token path with amount > 0.
// If positive is false, it returns all the token path.
func TokenList(positive bool) []string {
	tokens := []string{}

	for tokenPath, amount := range tokenListWithAmount {
		if positive && amount == 0 {
			continue
		}

		tokens = append(tokens, tokenPath)
	}

	return tokens
}

// GetTokenListWithAmount returns the token path and amount.
func GetTokenListWithAmount() map[string]uint64 {
	return tokenListWithAmount
}

// GetAmountOfToken returns the amount of token.
func GetAmountOfToken(tokenPath string) uint64 {
	amount, exists := tokenListWithAmount[tokenPath]
	if !exists {
		return 0
	}

	return amount
}

// GetAccuTransferToGovStaker returns the accuToGovStaker.
func GetAccuTransferToGovStaker() *avl.Tree {
	return accuToGovStaker
}

// GetAccuTransferToDevOps returns the accuToDevOps.
func GetAccuTransferToDevOps() *avl.Tree {
	return accuToDevOps
}

// GetAccuTransferToGovStakerByTokenPath returns the accumulated transfer to gov/staker by token path.
func GetAccuTransferToGovStakerByTokenPath(path string) uint64 {
	amountI, exists := accuToGovStaker.Get(path)
	if !exists {
		return 0
	}

	return amountI.(uint64)
}

// GetAccuTransferToDevOpsByTokenPath returns the accumulated transfer to devOps by token path.
func GetAccuTransferToDevOpsByTokenPath(path string) uint64 {
	amountI, exists := accuToDevOps.Get(path)
	if !exists {
		return 0
	}

	return amountI.(uint64)
}

// GetHistoryOfDistributedToGovStaker returns the history of distributed to gov/staker.
func GetHistoryOfDistributedToGovStaker() *avl.Tree {
	return distributedToGovStakerHistory
}

// GetHistoryOfDistributedToDevOps returns the history of distributed to devOps.
func GetHistoryOfDistributedToDevOps() *avl.Tree {
	return distributedToDevOpsHistory
}

// GetHistoryOfDistributedToGovStakerByTokenPath returns the history of distributed to gov/staker by token path.
func GetHistoryOfDistributedToGovStakerByTokenPath(path string) uint64 {
	amountI, exists := distributedToGovStakerHistory.Get(path)
	if !exists {
		return 0
	}

	return amountI.(uint64)
}

// GetHistoryOfDistributedToDevOpsByTokenPath returns the history of distributed to devOps by token path.
func GetHistoryOfDistributedToDevOpsByTokenPath(path string) uint64 {
	amountI, exists := distributedToDevOpsHistory.Get(path)
	if !exists {
		return 0
	}

	return amountI.(uint64)
}
