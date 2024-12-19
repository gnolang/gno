package users

import "errors"

var ErrInvalidUsername = errors.New("r/gnoland/users: invalid username")
var ErrNotRegistered = errors.New("r/gnoland/users: username not registered")
