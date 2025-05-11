package gnolang

import (
	"fmt"
	"regexp"

	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	// NOTE: These are further restrictions upon the validation that already happens by std.MemPackage.Validate().
	// sub.domain.com/a/any
	// sub.domain.com/b/single
	// sub.domain.com/c/letter
	// sub.domain.com/d/works
	// sub.domain.com/r/realm
	// sub.domain.com/r/realm/path
	// sub.domain.com/p/package/path
	// See also tm2/pkg/std/memfile.go.
	reGnoPkgPathURL = regexp.MustCompile(`^([a-z0-9-]+\.)*[a-z0-9-]+\.[a-z]{2,}\/(?:[a-z])(?:\/_?[a-z][a-z0-9_]*)+$`)
	reGnoPkgPathStd = regexp.MustCompile(`^([a-z][a-z0-9_]*\/)*[a-z][a-z0-9_]+$`)
)

func ValidateMemPackage(mpkg *std.MemPackage) error {
	err := mpkg.ValidateBasic()
	if err != nil {
		return err
	}
	if true && // none of these match...
		!reGnoPkgPathURL.MatchString(mpkg.Path) &&
		!reGnoPkgPathStd.MatchString(mpkg.Path) {
		return fmt.Errorf("invalid package/realm path %q", mpkg.Path)
	}
	return nil
}
