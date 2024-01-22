package gnoclient

import "errors"

var (
	errInvalidPkgPath  = errors.New("invalid pkgpath")
	errInvalidFuncName = errors.New("invalid function name")
)
