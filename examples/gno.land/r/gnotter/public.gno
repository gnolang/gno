package gnotter

import "std"

func CreateProfile(name string) ProfileID {
	std.AssertOriginCall()
	profileId := incProfileID()
	caller := std.GetOrigCaller()
	url := "/r/profile:" + name
	profile := newProfile(profileId, url, name, caller)
	profileIdKey := profileIDKey(profileId)
	gProfiles.Set(profileIdKey, profile)
	gProfilesByName.Set(name, profile)
	return profile.id
}

func GetProfileIDFromName(name string) (ProfileID, bool) {
	profileI, exists := gProfilesByName.Get(name)
	if !exists {
		return 0, false
	}
	return profileI.(*Profile).id, true
}

func Follow(followerID ProfileID, followingID ProfileID) ProfileID {
	std.AssertOriginCall()
	caller := std.GetOrigCaller()
	followerProfile := getProfile(followerID)
	followerProfile.Follow(followingID)
	return followingID
}
