package vm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// gnomodFileName is normalized before hashing. See PackageContentHash.
const gnomodFileName = "gnomod.toml"

// keeperOwnedGnomod resets every gnomod.toml field the keeper writes at submit
// to the value both sides of an approval can derive on their own.
//
// It is the inverse of stampGnomod, and the two have to agree exactly: the
// approver hashes the source directory, the keeper hashes the blob it stamped,
// and any field the keeper writes that this does not reset makes the two hashes
// disagree forever. That is not hypothetical -- stampGnomod also rewrites
// Module, which the first version of this normalization missed, and a package
// whose gnomod.toml declared a module other than its deploy path could then
// never be enabled by anyone.
//
// So stampGnomod calls this rather than repeating the list. Adding a
// keeper-owned field is one edit here, and both sides pick it up.
// TestPackageContentHashSurvivesTheRealStamp is what holds them together.
func keeperOwnedGnomod(gm *gnomod.File, pkgPath string) {
	// The deploy path wins over whatever the author declared; nothing on the
	// submit path requires the two to match.
	gm.Module = pkgPath
	// Keeper bookkeeping in a file the submitter authors. Reset wholesale
	// rather than field by field: anything left standing here is
	// attacker-supplied. See MsgEnablePackage.PkgHeight for what covers the
	// values themselves, which an approver's local copy cannot know.
	gm.AddPkg = gnomod.AddPkg{}
}

// parseGnomodForHash parses the gnomod.toml that PackageContentHash covers.
//
// Deliberately not gnomod.ParseMemPackage, which falls back to the deprecated
// gno.mod. checkGnomodConstraints refuses gno.mod outright, so a package
// carrying one can never be stored -- but the fallback still parses it here,
// and then the substitution below finds no file named gnomod.toml and hashes
// the raw gno.mod bytes instead. The result is a confident hash over a file the
// chain will never hold. Refusing it names the real problem, and `gno mod tidy`
// is the fix.
func parseGnomodForHash(mpkg *std.MemPackage) (*gnomod.File, error) {
	f := mpkg.GetFile(gnomodFileName)
	if f == nil {
		return nil, fmt.Errorf("%s: no %s (run 'gno mod tidy' if this package still carries gno.mod)",
			mpkg.Path, gnomodFileName)
	}
	return gnomod.ParseBytes(f.Name, []byte(f.Body))
}

// PackageContentHash identifies the source an approver signed off on.
//
// Approval names a path, and a path's contents can change: the same creator may
// replace parked bytes at any time, and that replacement is also the legitimate
// retry after a failed enable. Without naming the bytes, an approver who read
// GOOD can be made to activate EVIL.
//
// gnomod.toml is parsed and re-encoded canonically after keeperOwnedGnomod
// resets what the keeper stamps at submit. That makes the submitter's copy and
// the stored copy agree while still binding the fields the AUTHOR declares --
// private, draft, ignore, replace and the gno version -- which excluding the
// file outright did not. A realm approved as private could otherwise be
// re-parked as public with the same .gno files and the same hash, and
// checkGnomodConstraints does not catch it on a first deployment: its guard is
// `priorPrivate && !gm.Private`, and priorPrivate is false when nothing is live
// at the path yet.
//
// What it does NOT bind is the [addpkg] section itself, and that is a limit of
// the -pkgdir flow rather than an oversight: the approver's local copy has no
// [addpkg] section to compare against, because the keeper writes it. A re-park
// with byte-identical sources therefore keeps this hash while changing creator,
// height and max_deposit. MsgEnablePackage.PkgHeight is what closes that.
//
// Each field is length-prefixed so that adjacent names and bodies cannot be
// re-cut to collide -- without it, a file "ab" holding "c" and a file "a"
// holding "bc" would hash alike.
//
// Returns an error rather than an empty hash: the empty string is also what a
// message carrying no approval at all decodes to, and both CLI callers would
// otherwise sign one.
func PackageContentHash(mpkg *std.MemPackage) (string, error) {
	gm, err := parseGnomodForHash(mpkg)
	if err != nil {
		return "", err
	}
	return packageContentHash(mpkg, *gm), nil
}

// packageContentHash is PackageContentHash for a caller that has already parsed
// the file, so an enable does not decode the same blob twice.
//
// gm is taken BY VALUE: keeperOwnedGnomod mutates it, and EnablePackage needs
// its own copy intact afterwards -- it reads AddPkg.Creator to decide
// OriginCaller. The value copy is enough because every field reset here is a
// value type; Replace is shared with the caller and never written.
func packageContentHash(mpkg *std.MemPackage, gm gnomod.File) string {
	keeperOwnedGnomod(&gm, mpkg.Path)
	canonical := gm.WriteString()

	files := make([]*std.MemFile, 0, len(mpkg.Files))
	for _, f := range mpkg.Files {
		if f.Name == gnomodFileName {
			f = &std.MemFile{Name: f.Name, Body: canonical}
		}
		files = append(files, f)
	}
	slices.SortFunc(files, func(a, b *std.MemFile) int {
		return strings.Compare(a.Name, b.Name)
	})

	h := sha256.New()
	for _, f := range files {
		fmt.Fprintf(h, "%d:%s%d:%s", len(f.Name), f.Name, len(f.Body), f.Body)
	}
	return hex.EncodeToString(h.Sum(nil))
}
