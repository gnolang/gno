// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_edit_thread_03_filetest

package z_edit_thread_03_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.EditThread(cross(cur), 404, 1, "Foo", "bar")
}

// Error:
// board does not exist with ID: 404
