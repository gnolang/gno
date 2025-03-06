package pkgerr

import (
	"errors"
	"std"

	"gno.land/p/leon/pkgerr"
)

const someGlobalErrMsg = "some global err"

var path = std.CurrentRealm().PkgPath()

func ReturnAnError() error {
	return pkgerr.New(errors.New(someGlobalErrMsg))
}
