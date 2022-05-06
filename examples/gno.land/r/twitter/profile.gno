package twitter

import (
	"gno.land/p/avl"
	"std"
)

type ProfileID uint64

type Profile struct {
	id        ProfileID
	url       string
	name      string
	creator   std.Address
	threads   *avl.MutTree
	postsCtr  uint64
	createdAt std.Time
}

func newProfile(id ProfileID, url string, name string, creator std.Address) *Profile {
	exists := gProfilesByName.Has(name)
	if exists {
		panic("profile already exists")
	}
	return &Profile{
		id:        id,
		url:       url,
		name:      name,
		creator:   creator,
		threads:   avl.NewMutTree(),
		createdAt: std.GetTimestamp(),
	}
}
