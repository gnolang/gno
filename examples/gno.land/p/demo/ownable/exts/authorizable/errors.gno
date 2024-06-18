package authorizable

import "errors"

var (
	ErrNotInAuthList = errors.New("unauthorized; caller is not in authorized list")
	ErrNotSuperuser  = errors.New("unauthorized; caller is not superuser")
	ErrAlreadyInList = errors.New("address is already in authorized list")
)
