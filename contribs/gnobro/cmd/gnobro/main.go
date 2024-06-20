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
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
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

}

func execBrowser(cfg *broCfg, args []string, io commands.IO) error {
	m := zone.New()
	// if len(os.Args) != 2 {
	// 	fmt.Fprintln(os.Stderr, "no nameOrBench32 given")
	// 	return
	// }

	remoteAddr := resolveUnixOrTCPAddr(cfg.target)
	cl, err := client.NewHTTPClient(remoteAddr)
	if err != nil {
		return fmt.Errorf("unable to create http client for %q: %w", remoteAddr, err)
	}

	broclient := gnoclient.Client{
		RPCClient: cl,
	}

	// name := os.Args[1]

	// kb, err := keys.NewKeyBaseFromDir(gnoenv.HomeDir())
	// if err != nil {
	// 	panic("unable to load keybase: " + err.Error())
	// }

	// if ok, err := kb.HasByNameOrAddress(name); !ok || err != nil {
	// 	if err != nil {
	// 		panic("invalid name: " + err.Error())
	// 	}

	// 	panic("unknown name/address: " + name)
	// }

	// fmt.Printf("[%s] Enter password: ", name)

	// password, err := terminal.ReadPassword(0)
	// if err != nil {
	// 	panic("error while reading password: " + err.Error())
	// }

	// if _, err := kb.ExportPrivKeyUnsafe(name, string(password)); err != nil {
	// 	panic("invalid password: " + err.Error())
	// }

	cmd := initCommandInput()
	p := tea.NewProgram(
		model{
			client:       &BroClient{client: &broclient},
			listFuncs:    newFuncList(),
			urlInput:     initURLInput(),
			commandInput: cmd,
			zone:         m,
		},
		tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("could not run program: %w", err)
	}

	return nil
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
