package governor

import "errors"

var (
	ErrDepositEnded     = errors.New("deposit ended")
	ErrProposalNotExist = errors.New("proposal not exist")
)
