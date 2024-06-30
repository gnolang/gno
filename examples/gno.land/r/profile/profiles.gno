package profile

import (
	"std"
	"strings"

	"gno.land/p/avl"
	"gno.land/p/ufmt"
	"gno.land/r/users"
)

// TODO: makes sense to implement an allowance system?

var profiles *avl.MutTree // std.Address.String() -> *Profile

func init() {
	// profiles = avl.NewMutTree()
}

func Update(dict map[string]interface{}) {
	// FIXME: ask a price per stored data length?
	currentUser := std.GetOrigCaller()
	profile := getOrCreateProfileByAddress(currentUser)
	profile.update(dict)
}

func GetByAddressOrName(aon users.AddressOrName) *Profile {
	addr := aon.Resolve()
	profile, found := profiles.Get(addr.String())
	if !found {
		return nil
	}
	return profile.(*Profile)
}

func getOrCreateProfileByAddress(addr std.Address) *Profile {
	// lookup existing profile
	profile, found := profiles.Get(addr.String())
	if found {
		return profile.(*Profile)
	}

	// create
	newProfile := &Profile{address: addr}
	profiles.Set(addr.String(), newProfile)
	return newProfile
}

func Render(path string) string {
	parts := strings.Split(path, "/")

	switch {
	case path == "":
		output := ufmt.Sprintf("stats: %d known profiles\n", profiles.Size())
		return output
	case len(parts) == 1:
		aon := users.AddressOrName(parts[0])
		profile := GetByAddressOrName(aon)
		if profile != nil {
			return profile.Render()
		}
		return "404: no such profile"
	}

	return "404: invalid URL"
}
