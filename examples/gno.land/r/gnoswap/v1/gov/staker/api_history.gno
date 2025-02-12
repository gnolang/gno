package staker

import (
	"std"
	"time"

	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"
)

// GetPossibleVotingAddressWithWeight returns the max voting weight + and possible voting address with weight
func GetPossibleVotingAddressWithWeight(endTimestamp uint64) (uint64, map[std.Address]uint64) {
	if endTimestamp > uint64(time.Now().Unix()) {
		panic(addDetailToError(
			errFutureTime,
			ufmt.Sprintf("endTimestamp(%d) > now(%d)", endTimestamp, time.Now().Unix()),
		))
	}

	totalWeight := uint64(0)
	addressWithWeight := make(map[std.Address]uint64)

	delegationSnapShotHistory.Iterate("", "", func(key string, value interface{}) bool {
		history := value.([]DelegationSnapShotHistory)
		toAddr := std.Address(key)

		for i := len(history) - 1; i >= 0; i-- {
			record := history[i]

			if record.updatedAt > endTimestamp {
				continue
			}

			addressWithWeight[toAddr] = record.amount
			totalWeight += record.amount

			break
		}

		return false
	})

	return totalWeight, addressWithWeight
}

// GetPossibleVotingAddressWithWeightJSON returns the max voting weight + and possible voting address with weight(string json format)
func GetPossibleVotingAddressWithWeightJSON(endTimestamp uint64) (uint64, string) {
	totalWeight, addressWithWeight := GetPossibleVotingAddressWithWeight(endTimestamp)

	possibleObj := json.ObjectNode("", nil)
	possibleObj.AppendObject("height", json.StringNode("height", ufmt.Sprintf("%d", std.GetHeight())))
	possibleObj.AppendObject("now", json.StringNode("now", ufmt.Sprintf("%d", time.Now().Unix())))

	possibleArr := json.ArrayNode("votingPower", nil)
	for to, weight := range addressWithWeight {
		addrWithWeightObj := json.ObjectNode("", nil)
		addrWithWeightObj.AppendObject("address", json.StringNode("address", to.String()))
		addrWithWeightObj.AppendObject("weight", json.StringNode("weight", ufmt.Sprintf("%d", weight)))
		possibleArr.AppendArray(addrWithWeightObj)
	}

	possibleObj.AppendObject("votingPower", possibleArr)

	return totalWeight, marshal(possibleObj)
}
