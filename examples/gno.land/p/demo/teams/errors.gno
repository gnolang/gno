package teams

import "errors"

var (
	ErrTeamAddressAlreadyExists	= errors.New("Team address already exists")
	ErrTeamNotFound 			= errors.New("team not found")
	ErrUserAlreadyMember 		= errors.New("user already a member")
	ErrUserNotMember 			= errors.New("user not a member")
)