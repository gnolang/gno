package access

import "errors"

var (
	ErrNotInSet       = errors.New("access: address not in set")
	ErrNoMembersInSet = errors.New("access: set is empty")
	ErrLastMember     = errors.New("access: cannot delete last member")
)
