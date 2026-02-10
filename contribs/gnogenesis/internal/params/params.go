package params

import (
	"errors"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
)

type paramsCfg struct {
	common.Cfg
}

var errInvalidGenesisStateType = errors.New("invalid genesis state type")

// The 'params' struct is used to create shorthand access to specific fields
// within the params structure, such as accessing `auth.params.tx_sig_limit`
// through `auth.tx_sig_limit`.
type params struct {
	Auth *auth.Params `json:"auth" comment:"##### Auth params #####"`
	VM   *vm.Params   `json:"vm" comment:"##### VM params #####"`
	Bank *bank.Params `json:"bank" comment:"##### Bank params #####"`
}

// NewParamsCmd creates the genesis params subcommand
func NewParamsCmd(io commands.IO) *commands.Command {
	cfg := &paramsCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "params",
			ShortUsage: "<subcommand> [flags]",
			ShortHelp:  "manages the initial genesis params",
			LongHelp:   "Manages genesis params fields, use `get -h` for more informations about available params",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newParamsSetCmd(cfg, io),
		newParamsGetCmd(cfg, io),
	)

	return cmd
}
