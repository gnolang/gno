package gnoverse

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/db"
)

func NewTestingSandbox() Sandbox {
	opts := SandboxOpts{}
	opts.DB = db.NewMemDB()
	sandbox, err := NewSandbox(opts)
	if err != nil {
		panic(fmt.Errorf("init testing sandbox: %w", err))
	}
	return sandbox
}
