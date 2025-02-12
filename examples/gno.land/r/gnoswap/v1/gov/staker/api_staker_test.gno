package staker

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/json"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

func TestGetLockedInfoByAddress(t *testing.T) {
	addr := testutils.TestAddress("locked_info_test")

	now := uint64(time.Now().Unix())
	lockedGNSList := []lockedGNS{
		{amount: 1000, unlock: now - 100}, // already unlocked
		{amount: 2000, unlock: now + 100}, // still locked
		{amount: 3000, unlock: now - 50},  // already unlocked
	}
	addrLockedGns.Set(addr.String(), lockedGNSList)

	result := GetLockedInfoByAddress(addr)

	node := json.Must(json.Unmarshal([]byte(result)))

	uassert.True(t, node.HasKey("height"))
	uassert.True(t, node.HasKey("now"))
	uassert.True(t, node.HasKey("totalLocked"))
	uassert.True(t, node.HasKey("claimableAmount"))

	// summation of all the locked quantities
	totalLocked, err := node.MustKey("totalLocked").GetString()
	uassert.NoError(t, err)
	uassert.Equal(t, totalLocked, "6000") // 1000 + 2000 + 3000 = 6000

	claimableAmount, err := node.MustKey("claimableAmount").GetString()
	uassert.NoError(t, err)
	uassert.Equal(t, claimableAmount, "4000") // 1000 + 3000 = 4000
}

func TestGetLockedInfoByAddress_NoLocks(t *testing.T) {
	addr := testutils.TestAddress("no_locks_test")

	// no quantities are locked here
	result := GetLockedInfoByAddress(addr)
	uassert.Equal(t, result, "")
}

func TestGetLockedInfoByAddress_EmptyLocks(t *testing.T) {
	addr := testutils.TestAddress("empty_locks_test")

	addrLockedGns.Set(addr.String(), []lockedGNS{})

	result := GetLockedInfoByAddress(addr)

	node := json.Must(json.Unmarshal([]byte(result)))
	uassert.True(t, node.HasKey("height"))
	uassert.True(t, node.HasKey("now"))
	uassert.True(t, node.HasKey("totalLocked"))
	uassert.True(t, node.HasKey("claimableAmount"))

	totalLocked, err := node.MustKey("totalLocked").GetString()
	uassert.NoError(t, err)
	uassert.Equal(t, totalLocked, "0")

	claimableAmount, err := node.MustKey("claimableAmount").GetString()
	uassert.NoError(t, err)
	uassert.Equal(t, claimableAmount, "0")
}

func TestGetClaimableRewardByAddress(t *testing.T) {
	addr := testutils.TestAddress("claimable_test")

	rewardState.addStake(uint64(std.GetHeight()), addr, 100, 0, nil)

	currentGNSBalance = 1000
	// userEmissionReward.Set(addr.String(), uint64(1000))

	currentProtocolFeeBalance["token1:token2"] = 500
	currentProtocolFeeBalance["token2:token3"] = 300
	//pfTree := avl.NewTree()
	//pfTree.Set("token1:token2", uint64(500))
	//pfTree.Set("token2:token3", uint64(300))
	//userProtocolFeeReward.Set(addr.String(), pfTree)

	result := GetClaimableRewardByAddress(addr)

	node := json.Must(json.Unmarshal([]byte(result)))

	uassert.True(t, node.HasKey("height"))
	uassert.True(t, node.HasKey("now"))

	emissionReward, err := node.MustKey("emissionReward").GetString()
	uassert.NoError(t, err)
	uassert.Equal(t, emissionReward, "1000")

	protocolFees := node.MustKey("protocolFees")
	uassert.True(t, protocolFees.IsArray())

	protocolFees.ArrayEach(func(i int, fee *json.Node) {
		tokenPath, err := fee.MustKey("tokenPath").GetString()
		uassert.NoError(t, err)

		amount, err := fee.MustKey("amount").GetString()
		uassert.NoError(t, err)

		switch tokenPath {
		case "token1:token2":
			uassert.Equal(t, amount, "500")
		case "token2:token3":
			uassert.Equal(t, amount, "300")
		default:
			t.Errorf("unexpected tokenPath: %s", tokenPath)
		}
	})

	// clear test data
	currentGNSBalance = 0
	currentProtocolFeeBalance = make(map[string]uint64)
	rewardState = NewRewardState()
}

func TestGetClaimableRewardByAddress_NoRewards(t *testing.T) {
	addr := testutils.TestAddress("no_reward_test")

	result := GetClaimableRewardByAddress(addr)
	uassert.Equal(t, result, "")
}
