package boards

import (
	"errors"
	"std"
)

const (
	PermissionBoardCreate  Permission = "board:create"
	PermissionThreadCreate            = "thread:create"
	PermissionThreadEdit              = "thread:edit"
	PermissionThreadDelete            = "thread:delete"
	PermissionThreadRepost            = "thread:repost"
	PermissionReplyCreate             = "reply:create"
	PermissionReplyEdit               = "reply:edit"
	PermissionReplyDelete             = "reply:delete"
)

// ErrUnauzorized indicates that user doesn't have a required permission.
var ErrUnauzorized = errors.New("unauthorized")

type (
	// Permission defines the type for permissions.
	Permission string

	// Role defines the type for user roles.
	Role string

	// Args is a list of generic arguments.
	Args []interface{}

	// Permissioner define an interface to for permissioned execution.
	Permissioner interface {
		// WithPermission calls a callback when a user has a specific permission.
		// It panics on error.
		WithPermission(std.Address, Permission, Args, func(Args))
	}
)

// TODO: Replace by a default permissioner (use sane and secure defaults)
type allowAnyPermissioner struct{}

func (allowAnyPermissioner) WithPermission(_ std.Address, _ Permission, a Args, cb func(Args)) { cb(a) }
