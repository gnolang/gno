package vm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// PackageContentHash identifies the source an approver signed off on.
//
// Approval names a path, and a path's contents can change: the same creator may
// replace parked bytes at any time, and that replacement is also the legitimate
// retry after a failed enable. Without naming the bytes, an approver who read
// GOOD can be made to activate EVIL.
//
// The digest names every file exactly as the submitter sent it, gnomod.toml
// included, because those are the bytes an approver reviews. AddPackage records
// it before stamping gnomod.toml, so a parked blob does not hash to it: compare
// against the recorded value, never a digest recomputed from storage.
//
// Each field is length-prefixed so that adjacent names and bodies cannot be
// re-cut to collide -- without it, a file "ab" holding "c" and a file "a"
// holding "bc" would hash alike.
func PackageContentHash(mpkg *std.MemPackage) string {
	files := slices.Clone(mpkg.Files)
	slices.SortFunc(files, func(a, b *std.MemFile) int {
		return strings.Compare(a.Name, b.Name)
	})

	h := sha256.New()
	for _, f := range files {
		fmt.Fprintf(h, "%d:%s%d:%s", len(f.Name), f.Name, len(f.Body), f.Body)
	}
	return hex.EncodeToString(h.Sum(nil))
}
