package users

import "errors"

var ErrInvalidUsername = errors.New("users: invalid username")
var ErrAlreadyExists = errors.New("users: username already exists")
