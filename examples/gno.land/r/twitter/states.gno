package twitter

import "gno.land/p/avl"

var gProfiles = avl.NewMutTree()       // id -> *Profile
var gProfileCtr int                    // increments Profile.id
var gProfilesByName = avl.NewMutTree() // name -> *Profile
