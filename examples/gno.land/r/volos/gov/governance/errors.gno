package governance

import "errors"

var (
	ErrInsufficientXVLS     = errors.New("insufficient xVLS to propose")
	ErrStillHasXVLS         = errors.New("cannot remove member: address still holds xVLS")
	ErrNoXVLS               = errors.New("cannot add member: address does not hold any xVLS")
	ErrVotingDeadlineNotMet = errors.New("voting deadline not met")
	ErrNotStaker            = errors.New("only staker can call this function")
)
