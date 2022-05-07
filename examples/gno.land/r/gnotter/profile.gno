package gnotter

import (
	"gno.land/p/avl"
	"std"
	"strconv"
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
	Followers *avl.MutTree
	Following *avl.MutTree
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
		Followers: avl.NewMutTree(),
		Following: avl.NewMutTree(),
	}
}

func (profile *Profile) RenderProfile() string {
	str := ""
	str += "Name: " + profile.name + "\n"
	str += "URL: " + profile.url + "\n"
	str += "Following: " + strconv.Itoa(profile.Following.Size()) + "\n"
	str += "Follower: " + strconv.Itoa(profile.Followers.Size())
	return str
}
