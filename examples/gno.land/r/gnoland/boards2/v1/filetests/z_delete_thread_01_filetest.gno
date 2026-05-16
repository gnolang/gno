// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_delete_thread_01_filetest

package z_delete_thread_01_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.DeleteThread(cross2(cur), 404, 1)
}

// Error:
// board does not exist with ID: 404
