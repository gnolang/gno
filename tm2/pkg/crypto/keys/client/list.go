package client

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
)

const (
	multisigMembersNone  = "none"
	multisigMembersShort = "short"
	multisigMembersFull  = "full"

	multisigMembersShortLimit = 3
)

type ListCfg struct {
	RootCfg *BaseCfg

	MultisigMembers string
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
	fs.StringVar(
		&c.MultisigMembers,
		"multisig-members",
		multisigMembersShort,
		"show multisig member public keys: none, short (first 3), or full",
	)
}

func execList(cfg *ListCfg, args []string, io commands.IO) error {
	if len(args) != 0 {
		return flag.ErrHelp
	}

	switch cfg.MultisigMembers {
	case multisigMembersNone, multisigMembersShort, multisigMembersFull:
	default:
		return fmt.Errorf("invalid -multisig-members value %q: must be none, short, or full", cfg.MultisigMembers)
	}

	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.Home)
	if err != nil {
		return err
	}

	infos, err := kb.List()
	if err == nil {
		printInfos(infos, cfg.MultisigMembers, io)
	}

	return err
}

func printInfos(infos []keys.Info, multisigMembersMode string, io commands.IO) {
	for i, info := range infos {
		keyname := info.GetName()
		keytype := info.GetType()
		keypub := info.GetPubKey()
		keyaddr := info.GetAddress()
		keypath, _ := info.GetPath()

		keypubDisplay := keypub.String()
		if keytype == keys.TypeMulti {
			keypubDisplay = crypto.PubKeyToBech32(keypub)
		}

		io.Printfln("%d. %s (%s) - addr: %v pub: %v path: %v",
			i, keyname, keytype, keyaddr, keypubDisplay, keypath)

		if keytype != keys.TypeMulti || multisigMembersMode == multisigMembersNone {
			continue
		}

		msPub, ok := keypub.(multisig.PubKeyMultisigThreshold)
		if !ok {
			continue
		}

		memberKeys := msPub.PubKeys
		limit := len(memberKeys)
		if multisigMembersMode == multisigMembersShort && limit > multisigMembersShortLimit {
			limit = multisigMembersShortLimit
		}

		for _, pk := range memberKeys[:limit] {
			io.Printfln("  %s", pk.String())
		}

		if multisigMembersMode == multisigMembersShort && len(memberKeys) > multisigMembersShortLimit {
			io.Printfln("  ... and %d more (use -multisig-members=full to see all)", len(memberKeys)-multisigMembersShortLimit)
		}
	}
}
