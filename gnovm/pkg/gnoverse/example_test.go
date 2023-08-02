package gnoverse_test

import (
	"fmt"

	"github.com/gnolang/gno/gnovm/pkg/gnoverse"
)

func Example() {
	opts := gnoverse.SandboxOpts{}
	opts.WithDiskStore("")
	sandbox, _ := gnoverse.NewSandbox(opts)
	_ = sandbox
	fmt.Println("Done.")
	// Output:
	// Done.
}
