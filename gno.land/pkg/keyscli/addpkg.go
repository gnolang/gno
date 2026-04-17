package keyscli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type MakeAddPkgCfg struct {
	RootCfg    *client.MakeTxCfg
	PkgPath    string
	PkgDir     string
	Send       string
	MaxDeposit string
	Force      bool
}

func NewMakeAddPkgCmd(rootCfg *client.MakeTxCfg, io commands.IO) *commands.Command {
	cfg := &MakeAddPkgCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "addpkg",
			ShortUsage: "addpkg [flags] <key-name>",
			ShortHelp:  "uploads a new package",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execMakeAddPkg(cfg, args, io)
		},
	)
}

func (c *MakeAddPkgCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.PkgPath,
		"pkgpath",
		"",
		"package path (required)",
	)

	fs.StringVar(
		&c.PkgDir,
		"pkgdir",
		"",
		"path to package files (required)",
	)

	fs.StringVar(
		&c.Send,
		"send",
		"",
		"send amount",
	)

	fs.StringVar(
		&c.MaxDeposit,
		"max-deposit",
		"",
		"max storage deposit",
	)

	fs.BoolVar(
		&c.Force,
		"force",
		false,
		"force deployment even if there is a large version gap (> 5)",
	)
}

func execMakeAddPkg(cfg *MakeAddPkgCfg, args []string, io commands.IO) error {
	if cfg.PkgPath == "" {
		return errors.New("pkgpath not specified")
	}
	if cfg.PkgDir == "" {
		return errors.New("pkgdir not specified")
	}
	if cfg.RootCfg.GasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}
	if cfg.RootCfg.GasFee == "" {
		return errors.New("gas-fee not specified")
	}

	if len(args) != 1 {
		return flag.ErrHelp
	}

	// read account pubkey.
	nameOrBech32 := args[0]
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	creator := info.GetAddress()
	// info.GetPubKey()
	// Parse send amount.
	send, err := std.ParseCoins(cfg.Send)
	if err != nil {
		return errors.Wrap(err, "parsing send coins")
	}
	// parse deposit.
	deposit, err := std.ParseCoins(cfg.MaxDeposit)
	if err != nil {
		panic(err)
	}

	// open files in directory as MemPackage.
	memPkg := gno.MustReadMemPackage(cfg.PkgDir, cfg.PkgPath, gno.MPUserAll)
	if memPkg.IsEmpty() {
		panic(fmt.Sprintf("found an empty package %q", cfg.PkgPath))
	}

	// Check for version gaps (soft warning / hard block for large gaps).
	if err := checkVersionGap(cfg, io); err != nil {
		return err
	}

	// parse gas wanted & fee.
	gaswanted := cfg.RootCfg.GasWanted
	gasfee, err := std.ParseCoin(cfg.RootCfg.GasFee)
	if err != nil {
		panic(err)
	}
	// construct msg & tx and marshal.
	msg := vm.MsgAddPackage{
		Creator:    creator,
		Package:    memPkg,
		Send:       send,
		MaxDeposit: deposit,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       cfg.RootCfg.Memo,
	}

	if cfg.RootCfg.Broadcast {
		cfg.RootCfg.RootCfg.OnTxSuccess = func(tx std.Tx, res *ctypes.ResultBroadcastTxCommit) {
			PrintTxInfo(tx, res, io)
		}
		err := client.ExecSignAndBroadcast(cfg.RootCfg, args, tx, io)
		if err != nil {
			if isCLAError(err) {
				return wrapCLAError(err, cfg.RootCfg.RootCfg.Remote, cfg.RootCfg.ChainID, nameOrBech32)
			}
			return err
		}
	} else {
		io.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}

// maxVersionGap is the maximum allowed version gap before --force is required.
const maxVersionGap = 5

// checkVersionGap queries the chain for version information and emits
// warnings or errors when deploying a versioned package with gaps.
// Network/RPC errors are silently ignored so offline usage is unaffected.
func checkVersionGap(cfg *MakeAddPkgCfg, io commands.IO) error {
	basePath, version, ok := gno.ParseVersionSuffix(cfg.PkgPath)
	if !ok || version == 0 {
		return nil // not a versioned path or first version — nothing to check
	}

	remote := cfg.RootCfg.RootCfg.Remote
	if remote == "" {
		return nil // no remote configured — skip check
	}

	cli, err := rpcclient.NewHTTPClient(remote)
	if err != nil {
		return nil // silently skip on connection errors
	}

	qres, err := cli.ABCIQuery(context.Background(), "vm/qlatestversion", []byte(basePath))
	if err != nil {
		return nil // silently skip on network errors
	}

	if qres.Response.Error != nil {
		return evalVersionGap(basePath, version, nil, cfg.Force, io)
	}

	var result vm.LatestVersionResult
	if err := json.Unmarshal(qres.Response.Data, &result); err != nil {
		return nil // silently skip on parse errors
	}

	return evalVersionGap(basePath, version, &result, cfg.Force, io)
}

// evalVersionGap evaluates whether to warn or block based on the version being
// deployed and the on-chain state. If result is nil, no versions exist on-chain.
func evalVersionGap(basePath string, version int, result *vm.LatestVersionResult, force bool, io commands.IO) error {
	shortName := basePath
	if idx := strings.LastIndex(basePath, "/"); idx >= 0 {
		shortName = basePath[idx+1:]
	}

	// Determine gap from latest on-chain version.
	var latestVersion int
	if result != nil {
		latestStr := result.Latest
		if len(latestStr) > 0 && latestStr[0] == 'v' {
			latestStr = latestStr[1:]
		}
		var err error
		latestVersion, err = strconv.Atoi(latestStr)
		if err != nil {
			return nil // silently skip on parse errors
		}
	} else {
		latestVersion = -1
	}

	gap := version - latestVersion
	if gap <= 1 {
		return nil // sequential — nothing to warn about
	}

	// Warn about missing predecessor.
	if result == nil {
		io.ErrPrintfln("Warning: deploying %s/v%d but no previous versions exist on-chain.", shortName, version)
	} else {
		io.ErrPrintfln("Warning: deploying %s/v%d but %s/v%d does not exist on-chain (latest: %s).", shortName, version, shortName, version-1, result.Latest)
	}

	// Block if the gap is too large (unless --force is set).
	if gap > maxVersionGap && !force {
		return fmt.Errorf(
			"version gap too large: deploying v%d but latest on-chain is %s (gap: %d, max allowed: %d). Use --force to override",
			version, latestDisplay(result), gap, maxVersionGap,
		)
	}

	return nil
}

func latestDisplay(result *vm.LatestVersionResult) string {
	if result == nil {
		return "none"
	}
	return result.Latest
}
