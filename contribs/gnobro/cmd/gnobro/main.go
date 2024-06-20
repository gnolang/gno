package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/log"
	zone "github.com/lrstanley/bubblezone"
)

type broCfg struct {
	target  string
	chainID string
}

var defaultBroOptions = broCfg{
	target:  "127.0.0.1:26657",
	chainID: "dev",
}

func main() {
	cfg := &broCfg{}

	stdio := commands.NewDefaultIO()
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "gnobro",
			ShortUsage: "gnobro [flags] [path ...]",
			ShortHelp:  "runs a cli browser.",
			LongHelp:   `run a cli browser`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execBrowser(cfg, args, stdio)
		})

	cmd.Execute(context.Background(), os.Args[1:])
}

func (c *broCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.target,
		"target",
		defaultBroOptions.target,
		"target gnoland address",
	)

	fs.StringVar(
		&c.chainID,
		"chainid",
		defaultBroOptions.chainID,
		"chainid",
	)

}

func execBrowser(cfg *broCfg, args []string, io commands.IO) error {
	home := gnoenv.HomeDir()

	logger := log.NewNoopLogger()
	var address string
	var kb keys.Keybase
	if len(args) > 0 && args[0] != "" {
		address = args[0]

		var err error
		kb, err = keys.NewKeyBaseFromDir(home)
		if err != nil {
			return fmt.Errorf("unable to load keybase: %w", err)
		}
	} else {
		// create a inmemory keybase
		kb = keys.NewInMemory()
		kb.CreateAccount(integration.DefaultAccount_Name, integration.DefaultAccount_Seed, "", "", 0, 0)
		address = integration.DefaultAccount_Name
	}

	signer, err := getSignerForAccount(io, address, kb, cfg)
	if err != nil {
		return fmt.Errorf("unable to get signer for account %q: %w", address, err)
	}

	remoteAddr := resolveUnixOrTCPAddr(cfg.target)
	cl, err := client.NewHTTPClient(remoteAddr)
	if err != nil {
		return fmt.Errorf("unable to create http client for %q: %w", remoteAddr, err)
	}

	gnocl := &gnoclient.Client{
		RPCClient: cl,
		Signer:    signer,
	}
	base := gnoclient.BaseTxCfg{
		GasFee:    "1000000ugnot",
		GasWanted: 2000000,
	}

	broclient := NewBroClient(logger, base, gnocl)

	cmd := initCommandInput()
	mod := model{
		client:       broclient,
		listFuncs:    newFuncList(),
		urlInput:     initURLInput(),
		commandInput: cmd,
		zone:         zone.New(),
		pageurls:     map[string]string{},
	}

	p := tea.NewProgram(mod,
		tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("could not run program: %w", err)
	}

	return nil
}

func getSignerForAccount(io commands.IO, address string, kb keys.Keybase, cfg *broCfg) (gnoclient.Signer, error) {
	var signer gnoclient.SignerFromKeybase

	signer.Keybase = kb
	signer.Account = address
	signer.ChainID = cfg.chainID // XXX: override this
	// 	ChainID:  chainid, // Chain ID for transaction signing

	if ok, err := kb.HasByNameOrAddress(address); !ok || err != nil {
		if err != nil {
			return nil, fmt.Errorf("invalid name: %w", err)
		}

		return nil, fmt.Errorf("unknown name/address: %q", address)
	}

	// try empty password first
	if _, err := kb.ExportPrivKeyUnsafe(address, ""); err != nil {
		prompt := fmt.Sprintf("[%.10s] Enter password:", address)
		signer.Password, err = io.GetPassword(prompt, true)
		if err != nil {
			return nil, fmt.Errorf("error while reading password: %w", err)
		}

		if _, err := kb.ExportPrivKeyUnsafe(address, string(signer.Password)); err != nil {
			return nil, fmt.Errorf("invalid password: %w", err)
		}
	}

	return signer, nil
}

func resolveUnixOrTCPAddr(in string) (out string) {
	var err error
	var addr net.Addr

	if strings.HasPrefix(in, "unix://") {
		in = strings.TrimPrefix(in, "unix://")
		if addr, err := net.ResolveUnixAddr("unix", in); err == nil {
			return fmt.Sprintf("%s://%s", addr.Network(), addr.String())
		}

		err = fmt.Errorf("unable to resolve unix address `unix://%s`: %w", in, err)
	} else { // don't bother to checking prefix
		in = strings.TrimPrefix(in, "tcp://")
		if addr, err = net.ResolveTCPAddr("tcp", in); err == nil {
			return fmt.Sprintf("%s://%s", addr.Network(), addr.String())
		}

		err = fmt.Errorf("unable to resolve tcp address `tcp://%s`: %w", in, err)
	}

	panic(err)
}
