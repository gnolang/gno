package antisquat

import "errors"

var (
	ErrNameEmpty        = errors.New("antisquat: name cannot be empty")
	ErrNameTooShort     = errors.New("antisquat: name too short")
	ErrNameTooLong      = errors.New("antisquat: name too long")
	ErrNameInvalidChars = errors.New("antisquat: invalid characters in name")
	ErrNameTaken        = errors.New("antisquat: name already registered")
	ErrNameReserved     = errors.New("antisquat: name is reserved")
	ErrRateLimited      = errors.New("antisquat: registration rate limit exceeded")
	ErrNotOwner         = errors.New("antisquat: caller is not the name owner")
	ErrNameExpired      = errors.New("antisquat: name expired beyond grace period")
	ErrNameNotFound     = errors.New("antisquat: name not found")
	ErrNameNotExpirable = errors.New("antisquat: name does not expire")
)
