package client

import (
	"github.com/gnolang/gno/pkgs/command"
)

func RunMain(cmd *command.Command, args []string) error {

	// pop first argument.
	arg0 := args[0]

	// switch on first argument.
	switch arg0 {
	case "add":
		return cmd.Run(addApp, args[1:], DefaultAddOptions)
	case "delete":
		return cmd.Run(deleteApp, args[1:], DefaultDeleteOptions)
	case "list":
		return cmd.Run(listApp, args[1:], DefaultListOptions)
	case "sign":
		return cmd.Run(signApp, args[1:], DefaultSignOptions)
	case "verify":
		return cmd.Run(verifyApp, args[1:], DefaultVerifyOptions)
	default:
		panic("unknown command " + arg0)
	}
}
