package staker

import (
	"std"
	"time"

	"gno.land/p/demo/json"
)

// GetLockedAmount gets the total locked amount of GNS.
func GetLockedAmount() uint64 {
	return lockedAmount
}

// GetLockedInfoByAddress gets the locked info of an address.
// - total locked amount
// - claimable amount
func GetLockedInfoByAddress(addr std.Address) string {
	lockeds, exist := addrLockedGns.Get(addr.String())
	if !exist {
		return ""
	}

	now := uint64(time.Now().Unix())

	totalLocked := uint64(0)
	claimableAmount := uint64(0)

	for _, locked := range lockeds.([]lockedGNS) {
		amount := locked.amount
		unlock := locked.unlock

		totalLocked += amount

		if now >= unlock {
			claimableAmount += amount
		}
	}

	data := json.Builder().
		WriteString("height", formatInt(std.ChainHeight())).
		WriteString("now", formatInt(time.Now().Unix())).
		WriteString("totalLocked", formatUint(totalLocked)).
		WriteString("claimableAmount", formatUint(claimableAmount)).
		Node()

	return marshal(data)
}

func GetClaimableRewardByAddress(addr std.Address) string {
	rewardState.finalize(collectAndGetGNSBalance(), updateAndGetProtocolFeeBalance())

	emissionReward, protocolFeeRewards := rewardState.CalculateReward(addr)

	if emissionReward == 0 && len(protocolFeeRewards) == 0 {
		return ""
	}

	data := json.Builder().
		WriteString("height", formatInt(std.ChainHeight())).
		WriteString("now", formatInt(time.Now().Unix())).
		WriteString("emissionReward", formatUint(emissionReward)).
		Node()

	if len(protocolFeeRewards) > 0 {
		pfArr := json.ArrayNode("", nil)
		for tokenPath, protocolFeeReward := range protocolFeeRewards {
			if protocolFeeReward > 0 {
				pfObj := json.Builder().
					WriteString("tokenPath", tokenPath).
					WriteString("amount", formatUint(protocolFeeReward)).
					Node()
				pfArr.AppendArray(pfObj)
			}
		}

		data.AppendObject("protocolFees", pfArr)
	}

	return marshal(data)
}
