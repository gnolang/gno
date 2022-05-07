package gnotter

import (
	"strconv"
	"strings"
)

func padZero(u64 uint64, length int) string {
	str := strconv.Itoa(int(u64))
	if len(str) >= length {
		return str
	} else {
		return strings.Repeat("0", length-len(str)) + str
	}
}

func profileIDKey(profileID ProfileID) string {
	return padZero(uint64(profileID), 10)
}

func incProfileID() ProfileID {
	gProfileCtr++
	return ProfileID(gProfileCtr)
}

func getProfile(profileID ProfileID) *Profile {
	profileIdKey := profileIDKey(profileID)
	profile_, exists := gProfiles.Get(profileIdKey)
	if !exists {
		return nil
	}
	board := profile_.(*Profile)
	return board
}
