package gnoverse

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/db"
)

type Sandbox struct {
	db db.DB
	// banker
	// machine
}

type SandboxOpts struct {
	DB db.DB
}

func (opts SandboxOpts) Validate() error {
	if opts.DB == nil {
		return fmt.Errorf("missing DB")
	}
	return nil
}

func NewSandbox(opts SandboxOpts) (Sandbox, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid opts: %w", err)
	}
	box := Sandbox{
		db: opts.DB,
	}
	return box, nil
}
