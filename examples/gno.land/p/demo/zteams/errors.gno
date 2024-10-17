package zteams

import "errors"

var (
	errNotOwner                 = errors.New("not the owner")
	errNotAuthorized            = errors.New("not authorized")
	errUserNotDisburser         = errors.New("cannot disburse")
	errTeamNotFound             = errors.New("team not found")
	errUserNotMember            = errors.New("user not a member")
	errUserAlreadyMember        = errors.New("user is already a member")
	errTeamAddressAlreadyExists = errors.New("team address already exists")
)
