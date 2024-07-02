package pausable

import "errors"

var ErrUnauthorized = errors.New("unauthorized; address is not owner")
