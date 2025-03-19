package staker

import (
	"encoding/base64"
	"std"
	"testing"
)

func TestIncentiveIdCompute(t *testing.T) {
	caller := std.Address("g1xyz")
	targetPoolPath := "pool1"
	rewardToken := "gns"
	startTimestamp := int64(12345)
	endTimestamp := int64(67890)
	height := int64(1000)

	expected := base64.StdEncoding.EncodeToString([]byte("g1xyz:pool1:gns:12345:67890:1000"))
	actual := incentiveIdCompute(caller, targetPoolPath, rewardToken, startTimestamp, endTimestamp, height)

	if actual != expected {
		t.Errorf("incentiveIdCompute() = %s; want %s", actual, expected)
	}
}

func TestIncentiveIdByTime(t *testing.T) {
	startTime := int64(12345)
	endTime := int64(67890)
	creator := std.Address("g1xyz")
	rewardToken := "gns"

	expected := "00000000000000012345:00000000000000067890:g1xyz:gns"
	actual := incentiveIdByTime(startTime, endTime, creator, rewardToken)

	if actual != expected {
		t.Errorf("incentiveIdByTime() = %s; want %s", actual, expected)
	}
}

func TestIncentiveIdByHeight(t *testing.T) {
	startHeight := int64(123)
	endHeight := int64(456)
	creator := std.Address("g1xyz")
	rewardToken := "gns"

	expectedByHeight := "00000000000000000123:00000000000000000456:g1xyz:gns"
	expectedByCreator := "g1xyz:00000000000000000123:00000000000000000456:gns"

	actualByHeight, actualByCreator := incentiveIdByHeight(startHeight, endHeight, creator, rewardToken)

	if actualByHeight != expectedByHeight {
		t.Errorf("incentiveIdByHeight() byHeightId = %s; want %s", actualByHeight, expectedByHeight)
	}
	if actualByCreator != expectedByCreator {
		t.Errorf("incentiveIdByHeight() byCreatorId = %s; want %s", actualByCreator, expectedByCreator)
	}
}
