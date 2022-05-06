package twitter

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
