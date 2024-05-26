package memeland

import "errors"

var ErrUnauthorized = errors.New("unauthorized; caller is not owner")
