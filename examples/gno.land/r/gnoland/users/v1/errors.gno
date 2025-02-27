package users

import (
	"errors"

	"gno.land/p/demo/ufmt"
)

var (
	ErrPaused          = errors.New("r/gnoland/users: paused")
	ErrInvalidPayment  = ufmt.Errorf("r/gnoland/users: you need to send exactly %s ugnot", registerPrice)
	ErrInvalidUsername = errors.New("r/gnoland/users: invalid username")
)
