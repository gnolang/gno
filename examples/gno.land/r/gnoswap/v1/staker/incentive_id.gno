package staker

import (
	"encoding/base64"
	"std"

	"gno.land/p/demo/ufmt"
)

// incentiveIdCompute generates a unique incentive ID by combining caller address,
// target pool path, reward token, and time/height information.
//
// The generated ID is a Base64-encoded string of the following format:
// "caller:targetPoolPath:rewardToken:startTimestamp:endTimestamp:height".
//
// Parameters:
// - caller (std.Address): The address of the caller initiating the incentive.
// - targetPoolPath (string): The target pool path for the incentive.
// - rewardToken (string): The reward token associated with the incentive.
// - startTimestamp (int64): The starting timestamp of the incentive.
// - endTimestamp (int64): The ending timestamp of the incentive.
// - height (int64): The blockchain height at which the incentive is created.
//
// Returns:
// - string: A Base64-encoded string representing the unique incentive ID.
//
// Example:
// Input: caller="g1xyz", targetPoolPath="pool1", rewardToken="gns", startTimestamp=12345, endTimestamp=67890, height=1000
// Output: "ZzF4eXo6cG9vbDE6Z25zOjEyMzQ1OjY3ODkwOjEwMDA="
func incentiveIdCompute(caller std.Address, targetPoolPath, rewardToken string, startTimestamp, endTimestamp, height int64) string {
	key := ufmt.Sprintf("%s:%s:%s:%d:%d:%d", caller.String(), targetPoolPath, rewardToken, startTimestamp, endTimestamp, height)

	encoded := base64.StdEncoding.EncodeToString([]byte(key))
	return encoded
}

// incentiveIdByTime generates a unique incentive ID based on time intervals,
// creator address, and reward token.
//
// The generated ID is a plain string in the following format:
// "startTime:endTime:creator:rewardToken".
//
// Parameters:
// - startTime (uint64): The starting time of the incentive.
// - endTime (uint64): The ending time of the incentive.
// - creator (std.Address): The address of the incentive creator.
// - rewardToken (string): The reward token associated with the incentive.
//
// Returns:
// - string: A plain string representing the incentive ID.
//
// Example:
// Input: startTime=12345, endTime=67890, creator="g1xyz", rewardToken="gns"
// Output: "000000000000012345:000000000000067890:g1xyz:gns"
func incentiveIdByTime(startTime, endTime int64, creator std.Address, rewardToken string) string {
	startTimeEncode := EncodeUint(uint64(startTime))
	endTimeEncode := EncodeUint(uint64(endTime))
	creatorEncode := creator.String()

	byTimeId := []byte(startTimeEncode)
	byTimeId = append(byTimeId, ':')
	byTimeId = append(byTimeId, endTimeEncode...)
	byTimeId = append(byTimeId, ':')
	byTimeId = append(byTimeId, creatorEncode...)
	byTimeId = append(byTimeId, ':')
	byTimeId = append(byTimeId, rewardToken...)

	return string(byTimeId)
}

// incentiveIdByHeight generates two unique incentive IDs based on height intervals,
// creator address, and reward token.
//
// The first ID (byHeightId) has the format:
// "startHeight:endHeight:creator:rewardToken".
//
// The second ID (byCreatorId) has the format:
// "creator:startHeight:endHeight:rewardToken".
//
// Parameters:
// - startHeight (uint64): The starting blockchain height of the incentive.
// - endHeight (uint64): The ending blockchain height of the incentive.
// - creator (std.Address): The address of the incentive creator.
// - rewardToken (string): The reward token associated with the incentive.
//
// Returns:
// - string: The first ID (byHeightId).
// - string: The second ID (byCreatorId).
//
// Example:
// Input: startHeight=123, endHeight=456, creator="g1xyz", rewardToken="gns"
// Output: "000000000000000123:000000000000000456:g1xyz:gns", "g1xyz:000000000000000123:000000000000000456:gns"
func incentiveIdByHeight(startHeight, endHeight int64, creator std.Address, rewardToken string) (string, string) {
	startHeightEncode := EncodeUint(uint64(startHeight))
	endHeightEncode := EncodeUint(uint64(endHeight))
	creatorEncode := creator.String()

	byHeightId := []byte(startHeightEncode)
	byHeightId = append(byHeightId, ':')
	byHeightId = append(byHeightId, endHeightEncode...)
	byHeightId = append(byHeightId, ':')
	byHeightId = append(byHeightId, creatorEncode...)
	byHeightId = append(byHeightId, ':')
	byHeightId = append(byHeightId, rewardToken...)

	byCreatorId := []byte(creatorEncode)
	byCreatorId = append(byCreatorId, ':')
	byCreatorId = append(byCreatorId, startHeightEncode...)
	byCreatorId = append(byCreatorId, ':')
	byCreatorId = append(byCreatorId, endHeightEncode...)
	byCreatorId = append(byCreatorId, ':')
	byCreatorId = append(byCreatorId, rewardToken...)

	return string(byHeightId), string(byCreatorId)
}