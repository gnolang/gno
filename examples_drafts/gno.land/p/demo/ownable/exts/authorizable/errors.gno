package authorizable

import "errors"

var (
	ErrNotInAuthList = errors.New("authorizable: caller is not in authorized list")
	ErrNotSuperuser  = errors.New("authorizable: caller is not superuser")
	ErrAlreadyInList = errors.New("authorizable: address is already in authorized list")
)
