// Package auth provides object-based authentication interfaces and helpers.
package auth

import (
	"errors"
	"path"
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
)

// Token represents an authentication token
type Token interface {
	// Source is used to identify the provenance of the token.
	// It is intented to be used to find the corresponding [AuthenticateFn], for example in registries.
	// It can be spoofed and should not be trusted.
	Source() std.Realm
}

// AuthenticateFn validates a token and returns the ID of the authenticated entity.
// It panics with [ErrInvalidToken] if the token is invalid.
//
// XXX: Should we add some kind of `Token.EntityID` method instead of returning the ID here?
type AuthenticateFn = func(autok Token) string

var (
	ErrInvalidToken = errors.New("invalid token")
)

// NamespacedEntityID generates safe entity IDs from a namespace and a sub-path.
// subPath can be user-provided and will be sanitized to prevent overriding the namespace.
func NamespacedEntityID(namespace, subPath string) string {
	subPath = path.Clean(subPath)
	if strings.HasPrefix(subPath, "..") {
		panic(ufmt.Errorf("invalid sub-path %q", subPath))
	}
	return path.Join("/", namespace, subPath)
}
