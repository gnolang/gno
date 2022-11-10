package groups

import (
	"regexp"

	"gno.land/p/demo/avl"
)

//----------------------------------------
// Realm (package) state

var (
	gGroups       = avl.NewMutTree() // id -> *Group
	gGroupsCtr    int                // increments Group.id
	gGroupsByName = avl.NewMutTree() // name -> *Group
)

//----------------------------------------
// Constants

var reName = regexp.MustCompile(`^[a-z]+[_a-z0-9]{2,29}$`)
