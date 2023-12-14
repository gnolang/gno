//go:build deprecated

package client

import "github.com/gnolang/gno/tm2/pkg/commands"

// Deprecated: NewRootCmd is deprecated and has been moved in
// `gno.land/pkg/keycli`. Please update your code to use the new package. This
// function now only serves as a placeholder and will panic if called to
// encourage migration.
func NewRootCmd(io commands.IO) *commands.Command {
	panic("NewRootCmd: has been deprecated, use `gno.land/pkg/keycli` instead")
}

// Deprecated: NewRootCmdWithBaseConfig as has been moved in
// `gno.land/pkg/keycli`. Please update your code to use the new package. This
// function now only serves as a placeholder and will panic if called to
// encourage migration.
func NewRootCmdWithBaseConfig(io commands.IO, cfg interface{}) *commands.Command {
	panic("NewRootCmdWithBaseConfig: has been deprecated, use `gno.land/pkg/keycli` instead")
}
