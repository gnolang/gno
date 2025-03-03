package pkgerr

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/leon/pkgerr"
)

var (
	currRealmLink = strings.TrimPrefix(std.CurrentRealm().PkgPath(), "gno.land/")
	err           = errors.New("err test")
	wrappedErr    = pkgerr.New(err)
)

func errorOut() error {
	return wrappedErr
}
