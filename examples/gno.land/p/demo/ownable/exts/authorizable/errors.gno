package authorizable

import "errors"

var (
	ErrNotInAuthList = errors.New("unauthorized; caller is not in authorized list")
	ErrAlreadyInList = errors.New("address is already in authorized list")
)
