package keyscli

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	rpcClient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"golang.org/x/term"
)

// NewCLACmd creates the `gnokey cla` command with sign/status subcommands.
func NewCLACmd(cfg *client.BaseCfg, io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "cla",
			ShortUsage: "cla <subcommand>",
			ShortHelp:  "manage CLA (Contributor License Agreement) acknowledgment",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newCLASignCmd(cfg, io),
		newCLAStatusCmd(cfg, io),
	)

	return cmd
}

// CLASignCfg is the config for `gnokey cla sign`.
type CLASignCfg struct {
	RootCfg *client.BaseCfg
	CLAUrl  string // Override URL (optional, otherwise queried from chain)
}

func (c *CLASignCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.CLAUrl, "url", "", "override CLA document URL (default: query from chain)")
}

func newCLASignCmd(rootCfg *client.BaseCfg, cmdIO commands.IO) *commands.Command {
	cfg := &CLASignCfg{RootCfg: rootCfg}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "sign",
			ShortUsage: "cla sign [--url <url>] --remote <rpc-url>",
			ShortHelp:  "fetch, display, and sign the CLA for a specific chain",
			LongHelp:   "Queries the chain for CLA document URL, fetches content, displays it, and prompts for agreement. The CLA hash is stored in config.toml per remote for use with addpkg.",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execCLASign(cfg, cmdIO)
		},
	)
}

func execCLASign(cfg *CLASignCfg, cmdIO commands.IO) error {
	// Remote is always required to identify which chain
	if cfg.RootCfg.Remote == "" {
		return errors.New("--remote is required (identifies which chain this CLA is for)")
	}

	// Get CLA URL - either from flag or by querying chain
	claURL := cfg.CLAUrl
	if claURL == "" {
		url, err := queryCLADocURL(cfg.RootCfg.Remote)
		if err != nil {
			return fmt.Errorf("failed to query CLA document URL: %w", err)
		}
		if url == "" {
			return errors.New("CLA not configured on chain (cla_doc_url is empty)")
		}
		claURL = url
		cmdIO.Println("CLA document URL:", claURL)
	}

	// Fetch CLA content
	content, err := FetchCLAContent(claURL)
	if err != nil {
		return err
	}

	// Display CLA
	cmdIO.Println("")
	cmdIO.Println("=== Contributor License Agreement ===")
	cmdIO.Println(content)
	cmdIO.Println("=====================================")
	cmdIO.Println("")

	// Check if interactive
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("CLA signing requires interactive terminal")
	}

	// Prompt for agreement
	response, err := cmdIO.GetString("Type 'agree' to accept the CLA:")
	if err != nil {
		return err
	}

	if strings.ToLower(strings.TrimSpace(response)) != "agree" {
		return errors.New("CLA not accepted: you must type 'agree'")
	}

	// Compute hash and store in config
	hash := ComputeCLAHash(content)

	config, err := LoadConfig(cfg.RootCfg.Home)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	config.SetCLAHash(cfg.RootCfg.Remote, hash)

	if err := SaveConfig(cfg.RootCfg.Home, config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	cmdIO.Println("")
	cmdIO.Println("CLA signed successfully.")
	cmdIO.Println("Remote:", cfg.RootCfg.Remote)
	cmdIO.Println("Hash:", hash)

	return nil
}

// queryCLADocURL queries the chain for the CLA document URL parameter.
func queryCLADocURL(remote string) (string, error) {
	cli, err := rpcClient.NewHTTPClient(remote)
	if err != nil {
		return "", errors.Wrap(err, "new http client")
	}

	qres, err := cli.ABCIQuery(context.Background(), "params/vm:p:cla_doc_url", nil)
	if err != nil {
		return "", errors.Wrap(err, "querying")
	}

	if qres.Response.Error != nil {
		return "", qres.Response.Error
	}

	// Response data is JSON-quoted string, trim quotes
	data := string(qres.Response.Data)
	data = strings.TrimPrefix(data, `"`)
	data = strings.TrimSuffix(data, `"`)

	return data, nil
}

// CLAStatusCfg is the config for `gnokey cla status`.
type CLAStatusCfg struct {
	RootCfg *client.BaseCfg
}

func (c *CLAStatusCfg) RegisterFlags(fs *flag.FlagSet) {}

func newCLAStatusCmd(rootCfg *client.BaseCfg, cmdIO commands.IO) *commands.Command {
	cfg := &CLAStatusCfg{RootCfg: rootCfg}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "status",
			ShortUsage: "cla status [--remote <rpc-url>]",
			ShortHelp:  "show current CLA signature status",
			LongHelp:   "Shows CLA signature status. With --remote, shows status for that chain. Without --remote, shows all signed CLAs.",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execCLAStatus(cfg, cmdIO)
		},
	)
}

func execCLAStatus(cfg *CLAStatusCfg, cmdIO commands.IO) error {
	config, err := LoadConfig(cfg.RootCfg.Home)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.RootCfg.Remote != "" {
		// Show status for specific remote
		hash := config.GetCLAHash(cfg.RootCfg.Remote)
		if hash == "" {
			cmdIO.Println("CLA not signed for", cfg.RootCfg.Remote)
			cmdIO.Println("Run 'gnokey cla sign --remote", cfg.RootCfg.Remote+"' to sign the CLA.")
			return nil
		}
		cmdIO.Println("CLA signed for", cfg.RootCfg.Remote)
		cmdIO.Println("Hash:", hash)
	} else {
		// Show all signed CLAs
		if len(config.Zones) == 0 {
			cmdIO.Println("No CLAs signed.")
			cmdIO.Println("Run 'gnokey cla sign --remote <rpc-url>' to sign a CLA.")
			return nil
		}

		cmdIO.Println("Signed CLAs:")
		for remote, zone := range config.Zones {
			if zone.CLAHash != "" {
				cmdIO.Println(" ", remote, "->", zone.CLAHash)
			}
		}
	}

	return nil
}

// FetchCLAContent retrieves CLA text from the given URL or local path.
// Supports http://, https://, file:// URLs, and local filesystem paths.
func FetchCLAContent(urlOrPath string) (string, error) {
	// Handle file:// URLs
	if strings.HasPrefix(urlOrPath, "file://") {
		path := strings.TrimPrefix(urlOrPath, "file://")
		content, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read CLA file %s: %w", path, err)
		}
		return string(content), nil
	}

	// Handle local filesystem paths (absolute or relative)
	if !strings.HasPrefix(urlOrPath, "http://") && !strings.HasPrefix(urlOrPath, "https://") {
		content, err := os.ReadFile(urlOrPath)
		if err != nil {
			return "", fmt.Errorf("failed to read CLA file %s: %w", urlOrPath, err)
		}
		return string(content), nil
	}

	// HTTP/HTTPS URLs
	resp, err := http.Get(urlOrPath) // nolint:gosec
	if err != nil {
		return "", fmt.Errorf("failed to fetch CLA from %s: %w", urlOrPath, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("CLA fetch failed from %s: HTTP %d", urlOrPath, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read CLA content: %w", err)
	}

	return string(body), nil
}

// ComputeCLAHash returns a truncated SHA-256 hash of CLA content.
// Uses first 16 hex chars (8 bytes) as a tradeoff: collision-resistant enough
// for CLA versioning while keeping transaction size small.
func ComputeCLAHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)[:16]
}

// LoadCLAHashForRemote reads the CLA hash for a specific remote from config.
// Returns empty string if not found.
func LoadCLAHashForRemote(home, remote string) string {
	config, err := LoadConfig(home)
	if err != nil {
		return ""
	}
	return config.GetCLAHash(remote)
}
