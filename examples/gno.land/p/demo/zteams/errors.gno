package zteams

import "errors"

var (
	ErrNotOwner               = errors.New("not the owner")
	ErrNotAuthorized          = errors.New("not authorized")
	ErrUserNotDeployer		  = errors.New("cannot deploy")
	ErrTeamNotFound           = errors.New("team not found")
	ErrUserNotMember          = errors.New("user not a member")
	ErrUserAlreadyMember      = errors.New("user is already a member")
	ErrTeamAddressAlreadyExists = errors.New("team address already exists")
)