// PKGPATH: filetests/extern/foo

package foo

import (
	"filetests/extern/foo/internal"
	"filetests/extern/foo/internal/aa"
)

func main() {
	println(internal.A, internal.B)
	println(aa.A)
}

// Output:
// 123 456
// 456
