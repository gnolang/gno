package users

import (
	"errors"

	"gno.land/p/nt/ufmt"
)

var (
	ErrNonUserCall     = errors.New("r/gnoland/users: non-user call")
	ErrPaused          = errors.New("r/gnoland/users: paused")
	ErrInvalidPayment  = ufmt.Errorf("r/gnoland/users: you need to send exactly %d ugnot", registerPrice)
	ErrInvalidUsername = errors.New("r/gnoland/users: invalid username")
)
