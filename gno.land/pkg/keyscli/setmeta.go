package keyscli

import (
	"context"
	"flag"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type MakeSetMetaCfg struct {
	RootCfg *client.MakeTxCfg

	Send    string
	PkgPath string
	Fields  commands.StringArr
}

func NewMakeSetMetaCmd(rootCfg *client.MakeTxCfg, io commands.IO) *commands.Command {
	cfg := &MakeSetMetaCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "setmeta",
			ShortUsage: "setmeta [flags] <key-name or address>",
			ShortHelp:  "sets or updates the metadata of a package",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execMakeSetMeta(cfg, args, io)
		},
	)
}

func (c *MakeSetMetaCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.Send,
		"send",
		"",
		"send amount",
	)

	fs.StringVar(
		&c.PkgPath,
		"pkgpath",
		"",
		"package path (required)",
	)

	fs.Var(
		&c.Fields,
		"fields",
		"metadata fields (required)",
	)
}

func execMakeSetMeta(cfg *MakeSetMetaCfg, args []string, io commands.IO) error {
	if cfg.PkgPath == "" {
		return errors.New("pkgpath not specified")
	}
	if len(cfg.Fields) == 0 {
		return errors.New("no metadata fields were specified")
	}
	if len(args) != 1 {
		return flag.ErrHelp
	}

	// Get caller account address
	nameOrBech32 := args[0]
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	caller := info.GetAddress()

	// Parse metadata fields
	var fields []*vm.MetaField
	for _, s := range cfg.Fields {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) != 2 {
			return errors.New("invalid metadata field format")
		}

		name := strings.TrimSpace(parts[0])
		if name == "" {
			return errors.New("empty metadata field name")
		}

		var value []byte
		if v := parts[1]; strings.TrimSpace(v) != "" {
			value = []byte(v)
		}

		fields = append(fields, &vm.MetaField{
			Name:  name,
			Value: value,
		})
	}

	// Parse send amount
	send, err := std.ParseCoins(cfg.Send)
	if err != nil {
		return errors.Wrap(err, "parsing send coins")
	}

	// Parse gas wanted & fee
	gaswanted := cfg.RootCfg.GasWanted
	gasfee, err := std.ParseCoin(cfg.RootCfg.GasFee)
	if err != nil {
		return errors.Wrap(err, "parsing gas fee coin")
	}

	// Create transaction
	msg := vm.MsgSetMeta{
		Caller:  caller,
		Send:    send,
		PkgPath: cfg.PkgPath,
		Fields:  fields,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       cfg.RootCfg.Memo,
	}

	if cfg.RootCfg.Broadcast {
		err := client.ExecSignAndBroadcast(cfg.RootCfg, args, tx, io)
		if err != nil {
			return err
		}
	} else {
		io.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}
