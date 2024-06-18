package authorizable

import "errors"

var (
	errNotInAuthList = errors.New("unauthorized; caller is not in authorized list")
	errNotSuperuser  = errors.New("unauthorized; caller is not superuser")
	errAlreadyInList = errors.New("address is already in authorized list")
)
