package profile

import (
	"std"

	"gno.land/r/users"
)

// TODO: makes sense to implement an allowance system?

var profiles *avl.MutTree // std.Address.String() -> *Profile

func Update(dict map[string]interface{}) {
	// FIXME: ask a price per stored data length?
	currentUser := std.GetOrigCaller()
	profile := getOrCreateProfileByAddress(currentUser)
	profile.update(dict)
}

func GetByAddressOrName(aon users.AddressOrName) *Profile {
	return profiles.Get(aon)
}

func getOrCreateProfileByAddress(aon users.AddressOrName) *Profile {
	// lookup existing profile
	profile := profiles.Get(aon)
	if profile != nil {
		return profile
	}

	// create
	addr := aon.Resolve()
	newProfile := &Profile{address: addr}
	profiles.Set(addr, newProfile)
	return newProfile
}

func Render(path string) string {
	parts := strings.Split(path, "/")

	switch {
	case path == "":
		output := ufmt.Sprintf("stats: %d known profiles\n", profiles.Size())
		return output
	case len(parts) == 1:
		profile := GetByAddressOrName(parts[0])
		if profile != nil {
			return profile.Render()
		}
		return "404: no such profile"
	}

	return "404: invalid URL"
}
