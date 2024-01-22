package gnoclient

import "errors"

var (
	errInvalidPkgPath   = errors.New("invalid pkgpath")
	errInvalidFuncName  = errors.New("invalid function name")
	errMissingSigner    = errors.New("missing Signer")
	errMissingRPCClient = errors.New("missing RPCClient")
)
