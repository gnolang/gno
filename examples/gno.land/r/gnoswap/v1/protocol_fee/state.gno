package protocol_fee

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"

	"gno.land/r/gnoswap/v1/common"
)

// ProtocolFeeState holds all the state variables for protocol fee management
type ProtocolFeeState struct {
	// by default, devOps will get 0% of the protocol fee (which means gov/staker will get 100% of the protocol fee)
	DevOpsPct uint64 // 0%

	// accumulated amount distributed to gov/staker by token path
	AccuToGovStaker *avl.Tree
	AccuToDevOps    *avl.Tree

	// distributedToDevOpsHistory and distributedToGovStakerHistory are used to keep track of the distribution history
	DistributedToGovStakerHistory *avl.Tree
	DistributedToDevOpsHistory    *avl.Tree

	TokenListWithAmount map[string]uint64 // tokenPath -> amount
}

// distributedToGovStakerHistory returns the history of distributed to gov/staker.
func distributedToGovStakerHistory() *avl.Tree {
	return protocolFeeState.DistributedToGovStakerHistory
}

// distributedToDevOpsHistory returns the history of distributed to devOps.
func distributedToDevOpsHistory() *avl.Tree {
	return protocolFeeState.DistributedToDevOpsHistory
}

// NewProtocolFeeState creates a new instance of ProtocolFeeState with initialized values
func NewProtocolFeeState() *ProtocolFeeState {
	return &ProtocolFeeState{
		DevOpsPct:                     0, // 0%
		AccuToGovStaker:               avl.NewTree(),
		AccuToDevOps:                  avl.NewTree(),
		DistributedToGovStakerHistory: avl.NewTree(),
		DistributedToDevOpsHistory:    avl.NewTree(),
		TokenListWithAmount:           make(map[string]uint64),
	}
}

// Global instance of the protocol fee state
var protocolFeeState = NewProtocolFeeState()

// distributeToDevOps distributes tokens to DevOps and updates related state
// amount should be greater than 0 (already checked in DistributeProtocolFee)
func (pfs *ProtocolFeeState) distributeToDevOps(token string, amount uint64) error {
	pfs.addAccuToDevOps(token, amount)
	pfs.updateDistributedToDevOpsHistory(token, amount)
	if err := common.Transfer(cross, token, devOpsAddr, int64(amount)); err != nil {
		return ufmt.Errorf("transfer failed: token(%s), amount(%d)", token, amount)
	}

	return nil
}

// distributeToGovStaker distributes tokens to Gov/Staker and updates related state
// amount should be greater than 0 (already checked in DistributeProtocolFee)
func (pfs *ProtocolFeeState) distributeToGovStaker(token string, amount uint64) error {
	pfs.addAccuToGovStaker(token, amount)
	pfs.updateDistributedToGovStakerHistory(token, amount)
	if err := common.Transfer(cross, token, govStakerAddr, int64(amount)); err != nil {
		return ufmt.Errorf("transfer failed: token(%s), amount(%d)", token, amount)
	}

	return nil
}

// setDevOpsPct sets the devOpsPct.
func (pfs *ProtocolFeeState) setDevOpsPct(pct uint64) {
	if pct > 10000 {
		panic(addDetailToError(
			errInvalidPct,
			ufmt.Sprintf("pct(%d) should not be bigger than 10000", pct),
		))
	}

	prevDevOpsPct := pfs.DevOpsPct
	pfs.DevOpsPct = pct

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"SetDevOpsPct",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"newPct", strconv.FormatUint(pct, 10),
		"prevPct", strconv.FormatUint(prevDevOpsPct, 10),
	)
}

// addAccuToGovStaker adds the amount to the accuToGovStaker by token path.
func (pfs *ProtocolFeeState) addAccuToGovStaker(tokenPath string, amount uint64) {
	before := pfs.GetAccuTransferToGovStakerByTokenPath(tokenPath)
	after := before + amount
	pfs.AccuToGovStaker.Set(tokenPath, after)
}

// addAccuToDevOps adds the amount to the accuToDevOps by token path.
func (pfs *ProtocolFeeState) addAccuToDevOps(tokenPath string, amount uint64) {
	before := pfs.GetAccuTransferToDevOpsByTokenPath(tokenPath)
	after := before + amount
	pfs.AccuToDevOps.Set(tokenPath, after)
}

// GetAccuTransferToGovStakerByTokenPath gets the accumulated amount to gov/staker by token path.
func (pfs *ProtocolFeeState) GetAccuTransferToGovStakerByTokenPath(tokenPath string) uint64 {
	amountI, exists := pfs.AccuToGovStaker.Get(tokenPath)
	if !exists {
		return 0
	}
	return amountI.(uint64)
}

// GetAccuTransferToDevOpsByTokenPath gets the accumulated amount to devOps by token path.
func (pfs *ProtocolFeeState) GetAccuTransferToDevOpsByTokenPath(tokenPath string) uint64 {
	amountI, exists := pfs.AccuToDevOps.Get(tokenPath)
	if !exists {
		return 0
	}
	return amountI.(uint64)
}

// updateDistributedToGovStakerHistory updates the distributedToGovStakerHistory.
func (pfs *ProtocolFeeState) updateDistributedToGovStakerHistory(tokenPath string, amount uint64) {
	// get the previous amount
	prevAmountI, exists := pfs.DistributedToGovStakerHistory.Get(tokenPath)
	prevAmount := uint64(0)
	if exists {
		prevAmount = prevAmountI.(uint64)
	}

	// update the amount
	afterAmount := prevAmount + amount
	pfs.DistributedToGovStakerHistory.Set(tokenPath, afterAmount)
}

// updateDistributedToDevOpsHistory updates the distributedToDevOpsHistory.
func (pfs *ProtocolFeeState) updateDistributedToDevOpsHistory(tokenPath string, amount uint64) {
	// get the previous amount
	prevAmountI, exists := pfs.DistributedToDevOpsHistory.Get(tokenPath)
	prevAmount := uint64(0)
	if exists {
		prevAmount = prevAmountI.(uint64)
	}

	// update the amount
	afterAmount := prevAmount + amount
	pfs.DistributedToDevOpsHistory.Set(tokenPath, afterAmount)
}

// clearTokenListWithAmount clears the tokenListWithAmount.
func (pfs *ProtocolFeeState) clearTokenListWithAmount() {
	pfs.TokenListWithAmount = make(map[string]uint64)
}
