package client

import "github.com/gnolang/gno/tm2/pkg/commands"

// Deprecated: NewRootCmd is deprecated and no longer maintained. It has been
// moved in `gno.land/pkg/keycli`. Please update your code to use the new
// package. This function now only serves as a placeholder and will panic if
// called to encourage migration.
func NewRootCmd(io commands.IO) *commands.Command {
	panic("NewRootCmd: has been deprecated, use `gno.land/pkg/keycli` instead")
}
