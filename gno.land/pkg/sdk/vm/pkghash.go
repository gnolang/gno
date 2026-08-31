package vm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// gnomodFileName is excluded from the content hash. See PackageContentHash.
const gnomodFileName = "gnomod.toml"

// PackageContentHash identifies the source an approver signed off on.
//
// Approval names a path, and a path's contents can change: the same creator may
// replace parked bytes at any time, and that replacement is also the legitimate
// retry after a failed enable. Without naming the bytes, an approver who read
// GOOD can be made to activate EVIL.
//
// gnomod.toml is excluded deliberately. AddPackage stamps the creator, height
// and declared deposit into it at submit, so the stored file differs from the
// one the submitter sent and from the one an approver saw in the transaction --
// the two could never agree on a hash that included it. stampGnomod writes that
// file and touches nothing else, which is what makes excluding it sufficient:
// every byte an approver reviews is still covered, and the gnomod rules are
// re-applied from the stored file at enable anyway.
//
// Each field is length-prefixed so that adjacent names and bodies cannot be
// re-cut to collide -- without it, a file "ab" holding "c" and a file "a"
// holding "bc" would hash alike.
func PackageContentHash(mpkg *std.MemPackage) string {
	files := make([]*std.MemFile, 0, len(mpkg.Files))
	for _, f := range mpkg.Files {
		if f.Name == gnomodFileName {
			continue
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
