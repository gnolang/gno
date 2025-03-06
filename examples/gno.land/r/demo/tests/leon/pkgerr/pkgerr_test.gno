package pkgerr

import (
	"std"
	"strings"
	"testing"
)

func TestErr(t *testing.T) {
	err := ReturnAnError()
	pkgPath := std.CurrentRealm().PkgPath()
	shouldGet := strings.TrimPrefix(pkgPath, "gno.land/") + ": " + someGlobalErrMsg

	if err.Error() != shouldGet {
		t.Errorf("expecting error to be %s, got %s", shouldGet, err.Error())
	}
}
