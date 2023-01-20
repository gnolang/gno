package people

import (
	"time"
	"gno.land/p/avl"
)

type PersonID string

type PersonType int  // not used yet.

type Person struct {
	PersonID
	CreatedAt time.Time
	PersonType
	Name string
}

// Look at Sage's impl of Groups, see if should/can be used.
type Team struct {
	Members avl.Tree
}
