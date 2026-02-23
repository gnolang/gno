package client

import (
	"context"
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

type ListCfg struct {
	RootCfg *BaseCfg

	ShowMultisigMembers bool
}

func NewListCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &ListCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "list",
			ShortUsage: "list",
			ShortHelp:  "lists all keys in the keybase",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execList(cfg, args, io)
		},
	)
}

func (c *ListCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.ShowMultisigMembers,
		"multisig-members",
		false,
		"show multisig member public keys instead of the multisig public key",
	)
}

func execList(cfg *ListCfg, args []string, io commands.IO) error {
	if len(args) != 0 {
		return flag.ErrHelp
	}

	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.Home)
	if err != nil {
		return err
	}

	infos, err := kb.List()
	if err == nil {
		printInfos(infos, cfg.ShowMultisigMembers, io)
	}

	return err
}

func printInfos(infos []keys.Info, showMultisigMembers bool, io commands.IO) {
	for i, info := range infos {
		keyname := info.GetName()
		keytype := info.GetType()
		keypub := info.GetPubKey()
		keyaddr := info.GetAddress()
		keypath, _ := info.GetPath()
		keypubDisplay := keypub.String()

		if keytype == keys.TypeMulti && !showMultisigMembers {
			keypubDisplay = crypto.PubKeyToBech32(keypub)
		}

		io.Printfln("%d. %s (%s) - addr: %v path: %v pub: %v",
			i, keyname, keytype, keyaddr, keypath, keypubDisplay)
	}
}
