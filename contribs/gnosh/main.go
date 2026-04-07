package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

const defaultRemote = "https://rpc.gno.land:443"
const defaultChainID = "portal-loop"

// baseCfg holds shared flags for all commands.
type baseCfg struct {
	remote  string
	chainID string
	home    string
	keyName string
	json    bool
	quiet   bool
}

func (c *baseCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.remote, "remote", defaultRemote, "remote RPC address")
	fs.StringVar(&c.chainID, "chainid", defaultChainID, "chain ID")
	fs.StringVar(&c.home, "home", defaultHome(), "gno config home directory")
	fs.StringVar(&c.keyName, "key", "", "key name or address from keybase")
	fs.BoolVar(&c.json, "json", false, "output as JSON")
	fs.BoolVar(&c.quiet, "quiet", false, "suppress non-essential output")
}

func defaultHome() string {
	if h := os.Getenv("GNOHOME"); h != "" {
		return h
	}
	if h := os.Getenv("GNO_HOME"); h != "" {
		return h
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.Getenv("HOME"), ".config", "gno")
	}
	return filepath.Join(dir, "gno")
}

// rpcClient creates an RPC client from the config.
func (c *baseCfg) rpcClient() (rpcclient.Client, error) {
	return rpcclient.NewHTTPClient(c.remote)
}

// keybase opens the keybase from the home directory.
func (c *baseCfg) keybase() (keys.Keybase, error) {
	return keys.NewKeyBaseFromDir(c.home)
}

// signer creates a gnoclient Signer from the keybase.
func (c *baseCfg) signer() (*gnoclient.SignerFromKeybase, error) {
	if c.keyName == "" {
		return nil, fmt.Errorf("--key is required for signing transactions")
	}
	kb, err := c.keybase()
	if err != nil {
		return nil, fmt.Errorf("opening keybase: %w", err)
	}
	return &gnoclient.SignerFromKeybase{
		Keybase:  kb,
		Account:  c.keyName,
		Password: "", // will prompt via IO
		ChainID:  c.chainID,
	}, nil
}

// client creates a gnoclient.Client with RPC only (no signer).
func (c *baseCfg) queryClient() (*gnoclient.Client, error) {
	rpc, err := c.rpcClient()
	if err != nil {
		return nil, err
	}
	return &gnoclient.Client{RPCClient: rpc}, nil
}

// signingClient creates a gnoclient.Client with both signer and RPC.
func (c *baseCfg) signingClient(io commands.IO) (*gnoclient.Client, error) {
	rpc, err := c.rpcClient()
	if err != nil {
		return nil, err
	}
	s, err := c.signer()
	if err != nil {
		return nil, err
	}
	// Prompt for password
	pass, err := io.GetPassword("Enter password:", true)
	if err != nil {
		return nil, fmt.Errorf("reading password: %w", err)
	}
	s.Password = pass
	return &gnoclient.Client{
		Signer:    s,
		RPCClient: rpc,
	}, nil
}

func main() {
	io := commands.NewDefaultIO()

	cfg := &baseCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "gnosh",
			ShortUsage: "gnosh <command> [flags] [args...]",
			ShortHelp:  "The gno shell — an opinionated CLI for interacting with gno.land chains.",
			LongHelp: `gnosh is a developer-friendly wrapper around gnokey that provides
auto-gas estimation, smart defaults, chainable output, and a better UX
for day-to-day gno chain interaction.

Commands:
  call      Execute a realm function (transaction)
  query     Read-only query (qeval or render)
  version   Print version information`,
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newCallCmd(cfg, io),
		newQueryCmd(cfg, io),
		newVersionCmd(io),
	)

	cmd.Execute(context.Background(), os.Args[1:])
}

// outputJSON marshals v as indented JSON and writes to io.
func outputJSON(io commands.IO, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	io.Println(string(data))
	return nil
}
