package keyscli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)


type networkDef struct {
	Name      string
	ChainID   string
	Remote    string
	GnowebURL string
	Desc      string
}

var knownNetworks = []networkDef{
	{Name: "dev", ChainID: "dev", Remote: "127.0.0.1:26657", GnowebURL: "http://127.0.0.1:8888", Desc: "local development"},
	{Name: "staging", ChainID: "staging", Remote: "https://rpc.staging.gno.land:443", GnowebURL: "https://staging.gno.land", Desc: "staging"},
	{Name: "gnoland1", ChainID: "gnoland1", Remote: "https://rpc.gno.land:443", GnowebURL: "https://gno.land", Desc: "betanet / main"},
	{Name: "test11", ChainID: "test11", Remote: "https://rpc.test11.testnets.gno.land:443", GnowebURL: "https://test11.testnets.gno.land", Desc: "test11"},
}

func findNetworkByChainID(chainID string) *networkDef {
	for i := range knownNetworks {
		if knownNetworks[i].ChainID == chainID {
			return &knownNetworks[i]
		}
	}
	return nil
}

func findNetworkByRemote(remote string) *networkDef {
	for i := range knownNetworks {
		if knownNetworks[i].Remote == remote {
			return &knownNetworks[i]
		}
	}
	return nil
}

// promptNetwork offers the known-networks list and returns (chainID, remote).
// The caller's current (chainID, remote) are offered as a "keep current"
// option. If the pair mismatches a known network, the user is warned.
func promptNetwork(io commands.IO, chainID, remote string) (string, string, error) {
	if chainID != "" && remote != "" {
		netByChain := findNetworkByChainID(chainID)
		netByRemote := findNetworkByRemote(remote)
		if netByChain != nil && netByRemote != nil && netByChain.ChainID != netByRemote.ChainID {
			io.ErrPrintfln("warning: chainid %q and remote %q belong to different known networks", chainID, remote)
		}
	}

	items := make([]commands.SelectItem, 0, len(knownNetworks)+2)
	if chainID != "" && remote != "" {
		items = append(items, commands.SelectItem{
			Name:        "keep",
			Description: fmt.Sprintf("keep current (%s @ %s)", chainID, remote),
		})
	}
	for _, n := range knownNetworks {
		items = append(items, commands.SelectItem{Name: n.ChainID, Description: n.Desc})
	}
	items = append(items, commands.SelectItem{Name: "manual", Description: "enter manually"})

	selected, err := commands.PromptSelect(io, "Chain ID:", items)
	if err != nil {
		return "", "", err
	}

	switch selected {
	case "keep":
		return chainID, remote, nil
	case "manual":
		cid, err := commands.PromptString(io, "Chain ID", "", requiredValidator("chain ID"))
		if err != nil {
			return "", "", err
		}
		rem, err := commands.PromptString(io, "Remote node", "", requiredValidator("remote"))
		if err != nil {
			return "", "", err
		}
		return cid, rem, nil
	}

	net := findNetworkByChainID(selected)
	if net == nil {
		return "", "", errors.New("unknown network: %s", selected)
	}
	return net.ChainID, net.Remote, nil
}


// promptKeyOrAddress lists the keybase entries and lets the user select one.
// If no keys exist, prompts for a bech32 address manually. If keys exist, the
// list includes an "address" escape hatch that opens a bech32 prompt.
func promptKeyOrAddress(kbHome string, io commands.IO) (string, error) {
	kb, err := keys.NewKeyBaseFromDir(kbHome)
	if err != nil {
		return "", err
	}

	klist, err := kb.List()
	if err != nil {
		return "", err
	}

	if len(klist) == 0 {
		return commands.PromptString(io, "Bech32 address", "", bech32Validator)
	}

	items := make([]commands.SelectItem, 0, len(klist)+1)
	for _, info := range klist {
		items = append(items, commands.SelectItem{
			Name:        info.GetName(),
			Description: info.GetAddress().String(),
		})
	}
	items = append(items, commands.SelectItem{Name: "address", Description: "enter a bech32 address"})

	selected, err := commands.PromptSelect(io, "Key:", items)
	if err != nil {
		return "", err
	}
	if selected == "address" {
		return commands.PromptString(io, "Bech32 address", "", bech32Validator)
	}
	return selected, nil
}

func resolveKeyAddress(kbHome, nameOrBech32 string) (string, error) {
	// Bech32 wins if it parses.
	if _, err := crypto.AddressFromBech32(nameOrBech32); err == nil {
		return nameOrBech32, nil
	}
	kb, err := keys.NewKeyBaseFromDir(kbHome)
	if err != nil {
		return "", err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return "", err
	}
	return info.GetAddress().String(), nil
}


func requiredValidator(fieldName string) func(string) error {
	return func(s string) error {
		if s == "" {
			return fmt.Errorf("%s is required", fieldName)
		}
		return nil
	}
}

func bech32Validator(s string) error {
	if s == "" {
		return errors.New("address is required")
	}
	if _, err := crypto.AddressFromBech32(s); err != nil {
		return fmt.Errorf("invalid bech32 address: %w", err)
	}
	return nil
}

func coinsValidator(s string) error {
	if s == "" {
		return nil
	}
	if _, err := std.ParseCoins(s); err != nil {
		return fmt.Errorf("invalid coins: %w", err)
	}
	return nil
}

func requiredCoinsValidator(fieldName string) func(string) error {
	return func(s string) error {
		if s == "" {
			return fmt.Errorf("%s is required", fieldName)
		}
		if _, err := std.ParseCoins(s); err != nil {
			return fmt.Errorf("invalid coins: %w", err)
		}
		return nil
	}
}

func coinValidator(s string) error {
	if s == "" {
		return errors.New("coin required")
	}
	if _, err := std.ParseCoin(s); err != nil {
		return fmt.Errorf("invalid coin: %w", err)
	}
	return nil
}

func gasIntValidator(s string) error {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid gas amount: %s", s)
	}
	if n <= 0 {
		return fmt.Errorf("invalid gas amount: must be positive")
	}
	return nil
}

func dirValidator(s string) error {
	if s == "" {
		return errors.New("directory is required")
	}
	info, err := os.Stat(s)
	if err != nil {
		return fmt.Errorf("cannot access directory: %w", err)
	}
	if !info.IsDir() {
		return errors.New("not a directory")
	}
	return nil
}

// promptRequired fills *dst by prompting when it's empty. Returns true iff a
// prompt was issued. When interactive prompting is disabled, returns the
// given error instead.
func promptRequired(io commands.IO, root *client.MakeTxCfg, dst *string, label, errMsg string, validate func(string) error) (bool, error) {
	if *dst != "" {
		return false, nil
	}
	if !canPrompt(root, io) {
		return false, errors.New(errMsg)
	}
	val, err := commands.PromptString(io, label, "", validate)
	if err != nil {
		return false, err
	}
	*dst = val
	return true, nil
}

// simulateGas runs the given tx through .app/simulate and returns gasUsed.
func simulateGas(remote string, tx std.Tx) (int64, error) {
	if remote == "" {
		return 0, errors.New("remote is empty")
	}
	cli, err := rpcclient.NewHTTPClient(remote)
	if err != nil {
		return 0, err
	}
	bz, err := amino.Marshal(tx)
	if err != nil {
		return 0, err
	}
	res, err := client.SimulateTx(cli, bz)
	if err != nil {
		return 0, err
	}
	if res.CheckTx.IsErr() {
		return 0, fmt.Errorf("simulation check failed: %s", res.CheckTx.Log)
	}
	if res.DeliverTx.IsErr() {
		return 0, fmt.Errorf("simulation failed: %s", res.DeliverTx.Log)
	}
	return res.DeliverTx.GasUsed, nil
}

// queryGasPrice queries auth/gasprice from the chain and returns the
// denomination and price per gas unit. Returns an error when unavailable.
func queryGasPrice(remote string) (std.GasPrice, error) {
	var gp std.GasPrice
	if remote == "" {
		return gp, errors.New("remote is empty")
	}
	cli, err := rpcclient.NewHTTPClient(remote)
	if err != nil {
		return gp, err
	}
	qres, err := cli.ABCIQuery(context.Background(), "auth/gasprice", []byte{})
	if err != nil {
		return gp, err
	}
	if err := amino.UnmarshalJSON(qres.Response.Data, &gp); err != nil {
		return gp, err
	}
	return gp, nil
}

// autoEstimateGasFee computes a gas fee string based on chain gas price.
// Falls back to "1ugnot" if the chain cannot answer.
func autoEstimateGasFee(remote string, gasWanted int64) string {
	gp, err := queryGasPrice(remote)
	if err != nil || gp.Gas == 0 {
		return "1ugnot"
	}
	fee := gasWanted/gp.Gas + 1
	fee = fee * gp.Price.Amount
	// 5% buffer
	fee = fee + (fee*5)/100
	return fmt.Sprintf("%d%s", fee, gp.Price.Denom)
}

// attachDummySig attaches a placeholder signature for the given signer so the
// tx passes ValidateBasic during simulation. Chain skips signature verification
// in simulate mode, but pubkey→address matching is still enforced.
func attachDummySig(tx *std.Tx, kbHome, nameOrBech32 string) error {
	kb, err := keys.NewKeyBaseFromDir(kbHome)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	tx.Signatures = []std.Signature{{
		PubKey:    info.GetPubKey(),
		Signature: make([]byte, 64),
	}}
	return nil
}

// handleGas resolves gas-wanted and gas-fee interactively.
// previewTx, when non-nil, is used to attempt auto-simulation. When simulation
// fails or isn't requested, falls back to manual prompts.
// Returns (gasAutoEst, gasEstUsed, error) — gasEstUsed is populated when
// simulation succeeded and was used.
func handleGas(cfg *client.MakeTxCfg, io commands.IO, keyName string, previewTx func() (std.Tx, error)) (bool, int64, error) {
	var (
		autoEst    bool
		gasEstUsed int64
	)
	needWanted := cfg.GasWanted == 0
	needFee := cfg.GasFee == ""
	if !needWanted && !needFee {
		return false, 0, nil
	}

	if needWanted {
		items := []commands.SelectItem{
			{Name: "auto", Description: "auto-estimate gas via simulation"},
			{Name: "manual", Description: "enter gas manually"},
		}
		selected, err := commands.PromptSelect(io, "Gas estimation:", items)
		if err != nil {
			return false, 0, err
		}
		if selected == "auto" && previewTx != nil {
			// Temporarily fill placeholder gas/fee so ante validation passes during simulate.
			origWanted, origFee := cfg.GasWanted, cfg.GasFee
			if cfg.GasWanted == 0 {
				cfg.GasWanted = 10_000_000
			}
			if cfg.GasFee == "" {
				cfg.GasFee = "1ugnot"
			}
			tx, err := previewTx()
			cfg.GasWanted, cfg.GasFee = origWanted, origFee
			if err == nil && keyName != "" {
				if sigErr := attachDummySig(&tx, cfg.RootCfg.Home, keyName); sigErr != nil {
					io.ErrPrintfln("gas estimation failed: %s", sigErr)
					err = sigErr
				}
			}
			if err == nil {
				used, simErr := simulateGas(cfg.RootCfg.Remote, tx)
				if simErr == nil && used > 0 {
					cfg.GasWanted = int64(float64(used) * 1.2)
					gasEstUsed = used
					autoEst = true
				} else if simErr != nil {
					io.ErrPrintfln("gas estimation failed: %s", simErr)
				}
			} else {
				io.ErrPrintfln("gas estimation failed: %s", err)
			}
		}
		if cfg.GasWanted == 0 {
			val, err := commands.PromptString(io, "Gas wanted", "", gasIntValidator)
			if err != nil {
				return false, 0, err
			}
			n, _ := strconv.ParseInt(val, 10, 64)
			cfg.GasWanted = n
		}
	}

	if needFee {
		// Try auto gas-fee based on chain price; fallback to manual prompt.
		var defaultFee string
		if autoEst {
			defaultFee = autoEstimateGasFee(cfg.RootCfg.Remote, cfg.GasWanted)
		} else {
			defaultFee = "1ugnot"
		}
		val, err := commands.PromptString(io, "Gas fee", defaultFee, coinValidator)
		if err != nil {
			return false, 0, err
		}
		cfg.GasFee = val
	}

	return autoEst, gasEstUsed, nil
}


type txSummary struct {
	Type       string
	KeyName    string
	KeyAddr    string
	PkgPath    string
	PkgDir     string
	SourcePath string
	FuncName   string
	Args       []string
	To         string
	Send       string
	MaxDeposit string
	GasWanted  int64
	GasFee     string
	GasAutoEst bool
	GasEstUsed int64
	Memo       string
	ChainID    string
	Remote     string
	GnowebURL  string
}

func printSummary(io commands.IO, s txSummary) {
	io.ErrPrintfln("  ─── Transaction Summary ───")
	io.ErrPrintfln("  Type:        %s", s.Type)
	io.ErrPrintfln("  Key:         %s (%s)", s.KeyName, s.KeyAddr)
	if s.To != "" {
		io.ErrPrintfln("  To:          %s", s.To)
	}
	if s.PkgPath != "" {
		io.ErrPrintfln("  Package:     %s", s.PkgPath)
	}
	if s.PkgDir != "" {
		io.ErrPrintfln("  Directory:   %s", s.PkgDir)
	}
	if s.SourcePath != "" {
		io.ErrPrintfln("  Source:      %s", s.SourcePath)
	}
	if s.FuncName != "" {
		io.ErrPrintfln("  Function:    %s", s.FuncName)
	}
	if len(s.Args) > 0 {
		io.ErrPrintfln("  Arguments:   %s", strings.Join(s.Args, ", "))
	}
	if s.Send != "" {
		io.ErrPrintfln("  Send:        %s", s.Send)
	} else {
		io.ErrPrintfln("  Send:        (none)")
	}
	if s.MaxDeposit != "" {
		io.ErrPrintfln("  Max Deposit: %s", s.MaxDeposit)
	} else {
		io.ErrPrintfln("  Max Deposit: (none)")
	}
	if s.GasAutoEst {
		io.ErrPrintfln("  Gas Wanted:  %d (auto-estimated: %d used × 1.2)", s.GasWanted, s.GasEstUsed)
	} else {
		io.ErrPrintfln("  Gas Wanted:  %d", s.GasWanted)
	}
	io.ErrPrintfln("  Gas Fee:     %s", s.GasFee)
	if s.Memo != "" {
		io.ErrPrintfln("  Memo:        %s", s.Memo)
	} else {
		io.ErrPrintfln("  Memo:        (none)")
	}
	io.ErrPrintfln("  Chain ID:    %s", s.ChainID)
	io.ErrPrintfln("  Remote:      %s", s.Remote)
	if s.GnowebURL != "" {
		io.ErrPrintfln("")
		io.ErrPrintfln("  View on gno.land:")
		io.ErrPrintfln("    %s", s.GnowebURL)
	}
	io.ErrPrintfln("  ────────────────────────────")
}


func printAirGapHints(io commands.IO, addr, chainID, remote, txPath, keyName, gnowebURL string) {
	io.ErrPrintfln("Saved unsigned tx to %s", txPath)
	io.ErrPrintfln("")
	io.ErrPrintfln("  ─── Air-Gap Signing Workflow ───")
	io.ErrPrintfln("  Step 1: Fetch account info (online machine)")
	io.ErrPrintfln("    $ gnokey query auth/accounts/%s -remote %s", addr, remote)
	io.ErrPrintfln("")
	io.ErrPrintfln("  Step 2: Sign the transaction (offline machine)")
	io.ErrPrintfln("    $ gnokey sign -tx-path %s \\", txPath)
	io.ErrPrintfln("        -chainid %s \\", chainID)
	io.ErrPrintfln("        -account-number <from step 1> \\")
	io.ErrPrintfln("        -account-sequence <from step 1> \\")
	io.ErrPrintfln("        %s", keyName)
	io.ErrPrintfln("")
	io.ErrPrintfln("  Step 3: Broadcast signed transaction (online machine)")
	io.ErrPrintfln("    $ gnokey broadcast -remote %s %s", remote, txPath)
	if gnowebURL != "" {
		io.ErrPrintfln("")
		io.ErrPrintfln("  View after broadcast:")
		io.ErrPrintfln("    %s", gnowebURL)
	}
	io.ErrPrintfln("  ─────────────────────────────────")
}

func defaultAirGapFilename(txType string) string {
	return "./unsigned_" + txType + ".tx"
}

func saveUnsignedTxAndPrintAirGap(tx std.Tx, cfg *client.MakeTxCfg, keyName, keyAddr string, io commands.IO, gnowebURL, txType string) error {
	defaultPath := defaultAirGapFilename(txType)
	txPath, err := commands.PromptString(io, "Save unsigned tx to", defaultPath, nil)
	if err != nil {
		return err
	}
	if txPath == "" {
		txPath = defaultPath
	}

	jsonBz := amino.MustMarshalJSON(tx)
	if err := os.WriteFile(txPath, jsonBz, 0o644); err != nil {
		return fmt.Errorf("failed to save unsigned tx: %w", err)
	}

	printAirGapHints(io, keyAddr, cfg.ChainID, cfg.RootCfg.Remote, txPath, keyName, gnowebURL)
	io.Println(string(jsonBz))
	return nil
}


func canPrompt(cfg *client.MakeTxCfg, io commands.IO) bool {
	return commands.IsIOInteractive(io) && !cfg.NoInteractive
}

func promptProceed(io commands.IO) (bool, error) {
	return io.GetConfirmation("Proceed?")
}

// promptOptionalString prompts for an optional string field. Enter returns "".
func promptOptionalString(io commands.IO, label string, validate func(string) error) (string, error) {
	return commands.PromptString(io, label+" (optional, Enter to skip)", "", validate)
}


func execMakeTxInteractive(cfg *client.MakeTxCfg, args []string, io commands.IO) error {
	txType, err := commands.PromptSelect(io, "Transaction type:", []commands.SelectItem{
		{Name: "send", Description: "sends native currency"},
		{Name: "addpkg", Description: "uploads a new package"},
		{Name: "call", Description: "executes a realm function call"},
		{Name: "run", Description: "runs Gno code by invoking main()"},
	})
	if err != nil {
		return err
	}

	// Offer the known-networks list in the full wizard. Subcommand flows
	// accept whatever the user passed via flags/defaults.
	chainID, remote, err := promptNetwork(io, cfg.ChainID, cfg.RootCfg.Remote)
	if err != nil {
		return err
	}
	cfg.ChainID = chainID
	cfg.RootCfg.Remote = remote

	keyName, err := promptKeyOrAddress(cfg.RootCfg.Home, io)
	if err != nil {
		return err
	}

	switch txType {
	case "send":
		sendCfg := &client.MakeSendCfg{RootCfg: cfg}
		return execMakeSendInteractive(sendCfg, []string{keyName}, io, true)
	case "addpkg":
		addCfg := &MakeAddPkgCfg{RootCfg: cfg}
		return execMakeAddPkgInteractive(addCfg, []string{keyName}, io, true)
	case "call":
		callCfg := &MakeCallCfg{RootCfg: cfg}
		return execMakeCallInteractive(callCfg, []string{keyName}, io, true)
	case "run":
		runCfg := &MakeRunCfg{RootCfg: cfg}
		return execMakeRunInteractive(runCfg, []string{keyName}, io, true)
	default:
		return errors.New("unknown transaction type: %s", txType)
	}
}


func execMakeCallInteractive(cfg *MakeCallCfg, args []string, io commands.IO, fullWizard bool) error {
	root := cfg.RootCfg
	prompted := false

	if len(args) < 1 {
		if !canPrompt(root, io) {
			return errors.New("key name or address required")
		}
		keyName, err := promptKeyOrAddress(root.RootCfg.Home, io)
		if err != nil {
			return err
		}
		args = []string{keyName}
		prompted = true
	}

	p, err := promptRequired(io, root, &cfg.PkgPath, "Package path", "pkgpath not specified", requiredValidator("pkgpath"))
	if err != nil {
		return err
	}
	prompted = prompted || p
	p, err = promptRequired(io, root, &cfg.FuncName, "Function name", "func not specified", requiredValidator("func"))
	if err != nil {
		return err
	}
	prompted = prompted || p

	if fullWizard {
		if len(cfg.Args) == 0 {
			val, err := promptOptionalString(io, "Arguments (comma-separated)", nil)
			if err != nil {
				return err
			}
			if val != "" {
				for _, p := range strings.Split(val, ",") {
					cfg.Args = append(cfg.Args, strings.TrimSpace(p))
				}
			}
		}
		if cfg.Send == "" {
			val, err := promptOptionalString(io, "Send amount", coinsValidator)
			if err != nil {
				return err
			}
			cfg.Send = val
		}
		if cfg.MaxDeposit == "" {
			val, err := promptOptionalString(io, "Max deposit", coinsValidator)
			if err != nil {
				return err
			}
			cfg.MaxDeposit = val
		}
	}

	autoEst, gasEstUsed, err := handleGas(root, io, args[0], func() (std.Tx, error) {
		return buildCallTx(cfg, args[0])
	})
	if err != nil {
		return err
	}
	if autoEst || root.GasWanted == 0 || root.GasFee == "" {
		prompted = true
	}

	if fullWizard && root.Memo == "" {
		val, err := promptOptionalString(io, "Memo", nil)
		if err != nil {
			return err
		}
		root.Memo = val
	}

	addr, _ := resolveKeyAddress(root.RootCfg.Home, args[0])
	gnowebURL := GnowebTxURL(root.RootCfg.Remote, cfg.PkgPath, cfg.FuncName)

	if prompted || !root.Broadcast {
		printSummary(io, txSummary{
			Type:       "call",
			KeyName:    args[0],
			KeyAddr:    addr,
			PkgPath:    cfg.PkgPath,
			FuncName:   cfg.FuncName,
			Args:       cfg.Args,
			Send:       cfg.Send,
			MaxDeposit: cfg.MaxDeposit,
			GasWanted:  root.GasWanted,
			GasFee:     root.GasFee,
			GasAutoEst: autoEst,
			GasEstUsed: gasEstUsed,
			Memo:       root.Memo,
			ChainID:    root.ChainID,
			Remote:     root.RootCfg.Remote,
			GnowebURL:  gnowebURL,
		})
		proceed, err := promptProceed(io)
		if err != nil {
			return err
		}
		if !proceed {
			io.ErrPrintfln("Cancelled.")
			return nil
		}
	}

	if !root.Broadcast && canPrompt(root, io) {
		tx, err := buildCallTx(cfg, args[0])
		if err != nil {
			return err
		}
		return saveUnsignedTxAndPrintAirGap(tx, root, args[0], addr, io, gnowebURL, "call")
	}

	return execMakeCall(cfg, args, io)
}

func buildCallTx(cfg *MakeCallCfg, nameOrBech32 string) (std.Tx, error) {
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.Home)
	if err != nil {
		return std.Tx{}, err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return std.Tx{}, err
	}
	send, err := std.ParseCoins(cfg.Send)
	if err != nil {
		return std.Tx{}, errors.Wrap(err, "parsing send coins")
	}
	deposit, err := std.ParseCoins(cfg.MaxDeposit)
	if err != nil {
		return std.Tx{}, errors.Wrap(err, "parsing storage deposit coins")
	}
	var gasfee std.Coin
	if cfg.RootCfg.GasFee != "" {
		gasfee, err = std.ParseCoin(cfg.RootCfg.GasFee)
		if err != nil {
			return std.Tx{}, errors.Wrap(err, "parsing gas fee coin")
		}
	}
	msg := vm.MsgCall{
		Caller:     info.GetAddress(),
		Send:       send,
		MaxDeposit: deposit,
		PkgPath:    cfg.PkgPath,
		Func:       cfg.FuncName,
		Args:       cfg.Args,
	}
	return std.Tx{
		Msgs: []std.Msg{msg},
		Fee:  std.NewFee(cfg.RootCfg.GasWanted, gasfee),
		Memo: cfg.RootCfg.Memo,
	}, nil
}


func execMakeAddPkgInteractive(cfg *MakeAddPkgCfg, args []string, io commands.IO, fullWizard bool) error {
	root := cfg.RootCfg
	prompted := false

	if len(args) < 1 {
		if !canPrompt(root, io) {
			return errors.New("key name or address required")
		}
		keyName, err := promptKeyOrAddress(root.RootCfg.Home, io)
		if err != nil {
			return err
		}
		args = []string{keyName}
		prompted = true
	}

	// Ask for the directory first, then derive the package path from
	// its gnomod.toml's `module = "..."` declaration. Users shouldn't
	// have to retype the path they already wrote in gnomod.toml.
	if cfg.PkgDir == "" {
		if !canPrompt(root, io) {
			return errors.New("pkgdir not specified")
		}
		cwd, _ := os.Getwd()
		val, err := commands.PromptString(io, "Package directory", cwd, dirValidator)
		if err != nil {
			return err
		}
		cfg.PkgDir = val
		prompted = true
	}
	if cfg.PkgPath == "" {
		if mod, err := gnomod.ParseDir(cfg.PkgDir); err == nil && mod.Module != "" {
			cfg.PkgPath = mod.Module
			io.ErrPrintfln("Package path (from gnomod.toml): %s", cfg.PkgPath)
		}
	}
	p, err := promptRequired(io, root, &cfg.PkgPath, "Package path", "pkgpath not specified", requiredValidator("pkgpath"))
	if err != nil {
		return err
	}
	prompted = prompted || p

	if fullWizard {
		if cfg.Send == "" {
			val, err := promptOptionalString(io, "Send amount", coinsValidator)
			if err != nil {
				return err
			}
			cfg.Send = val
		}
		if cfg.MaxDeposit == "" {
			val, err := promptOptionalString(io, "Max deposit", coinsValidator)
			if err != nil {
				return err
			}
			cfg.MaxDeposit = val
		}
	}

	autoEst, gasEstUsed, err := handleGas(root, io, args[0], nil)
	if err != nil {
		return err
	}
	if autoEst || root.GasWanted == 0 || root.GasFee == "" {
		prompted = true
	}

	if fullWizard && root.Memo == "" {
		val, err := promptOptionalString(io, "Memo", nil)
		if err != nil {
			return err
		}
		root.Memo = val
	}

	addr, _ := resolveKeyAddress(root.RootCfg.Home, args[0])
	gnowebURL := GnowebTxURL(root.RootCfg.Remote, cfg.PkgPath, "")

	if prompted || !root.Broadcast {
		printSummary(io, txSummary{
			Type:       "addpkg",
			KeyName:    args[0],
			KeyAddr:    addr,
			PkgPath:    cfg.PkgPath,
			PkgDir:     cfg.PkgDir,
			Send:       cfg.Send,
			MaxDeposit: cfg.MaxDeposit,
			GasWanted:  root.GasWanted,
			GasFee:     root.GasFee,
			GasAutoEst: autoEst,
			GasEstUsed: gasEstUsed,
			Memo:       root.Memo,
			ChainID:    root.ChainID,
			Remote:     root.RootCfg.Remote,
			GnowebURL:  gnowebURL,
		})
		proceed, err := promptProceed(io)
		if err != nil {
			return err
		}
		if !proceed {
			io.ErrPrintfln("Cancelled.")
			return nil
		}
	}

	if !root.Broadcast && canPrompt(root, io) {
		tx, err := buildAddPkgTx(cfg, args[0])
		if err != nil {
			return err
		}
		return saveUnsignedTxAndPrintAirGap(tx, root, args[0], addr, io, gnowebURL, "addpkg")
	}

	return execMakeAddPkg(cfg, args, io)
}


func execMakeRunInteractive(cfg *MakeRunCfg, args []string, io commands.IO, fullWizard bool) error {
	root := cfg.RootCfg
	prompted := false

	if len(args) < 1 {
		if !canPrompt(root, io) {
			return errors.New("key name or address required")
		}
		keyName, err := promptKeyOrAddress(root.RootCfg.Home, io)
		if err != nil {
			return err
		}
		args = []string{keyName}
		prompted = true
	}
	if len(args) < 2 {
		if !canPrompt(root, io) {
			return errors.New("source file or directory required")
		}
		val, err := commands.PromptString(io, "Source file or directory", "", requiredValidator("source path"))
		if err != nil {
			return err
		}
		args = append(args, val)
		prompted = true
	}

	if fullWizard {
		if cfg.Send == "" {
			val, err := promptOptionalString(io, "Send amount", coinsValidator)
			if err != nil {
				return err
			}
			cfg.Send = val
		}
		if cfg.MaxDeposit == "" {
			val, err := promptOptionalString(io, "Max deposit", coinsValidator)
			if err != nil {
				return err
			}
			cfg.MaxDeposit = val
		}
	}

	autoEst, gasEstUsed, err := handleGas(root, io, args[0], nil)
	if err != nil {
		return err
	}
	if autoEst || root.GasWanted == 0 || root.GasFee == "" {
		prompted = true
	}

	if fullWizard && root.Memo == "" {
		val, err := promptOptionalString(io, "Memo", nil)
		if err != nil {
			return err
		}
		root.Memo = val
	}

	addr, _ := resolveKeyAddress(root.RootCfg.Home, args[0])

	if prompted || !root.Broadcast {
		printSummary(io, txSummary{
			Type:       "run",
			KeyName:    args[0],
			KeyAddr:    addr,
			SourcePath: args[1],
			Send:       cfg.Send,
			MaxDeposit: cfg.MaxDeposit,
			GasWanted:  root.GasWanted,
			GasFee:     root.GasFee,
			GasAutoEst: autoEst,
			GasEstUsed: gasEstUsed,
			Memo:       root.Memo,
			ChainID:    root.ChainID,
			Remote:     root.RootCfg.Remote,
		})
		proceed, err := promptProceed(io)
		if err != nil {
			return err
		}
		if !proceed {
			io.ErrPrintfln("Cancelled.")
			return nil
		}
	}

	if !root.Broadcast && canPrompt(root, io) {
		tx, err := buildRunTx(cfg, args[0], args[1], io)
		if err != nil {
			return err
		}
		return saveUnsignedTxAndPrintAirGap(tx, root, args[0], addr, io, "", "run")
	}

	return execMakeRun(cfg, args, io)
}


func execMakeSendInteractive(cfg *client.MakeSendCfg, args []string, io commands.IO, fullWizard bool) error {
	root := cfg.RootCfg
	prompted := false

	if len(args) < 1 {
		if !canPrompt(root, io) {
			return errors.New("key name or address required")
		}
		keyName, err := promptKeyOrAddress(root.RootCfg.Home, io)
		if err != nil {
			return err
		}
		args = []string{keyName}
		prompted = true
	}

	p, err := promptRequired(io, root, &cfg.Send, "Send amount", "send (amount) must be specified", requiredCoinsValidator("send amount"))
	if err != nil {
		return err
	}
	prompted = prompted || p
	p, err = promptRequired(io, root, &cfg.To, "Destination address", "to (destination address) must be specified", bech32Validator)
	if err != nil {
		return err
	}
	prompted = prompted || p

	autoEst, gasEstUsed, err := handleGas(root, io, args[0], nil)
	if err != nil {
		return err
	}
	if autoEst || root.GasWanted == 0 || root.GasFee == "" {
		prompted = true
	}

	if fullWizard && root.Memo == "" {
		val, err := promptOptionalString(io, "Memo", nil)
		if err != nil {
			return err
		}
		root.Memo = val
	}

	addr, _ := resolveKeyAddress(root.RootCfg.Home, args[0])

	if prompted || !root.Broadcast {
		printSummary(io, txSummary{
			Type:       "send",
			KeyName:    args[0],
			KeyAddr:    addr,
			To:         cfg.To,
			Send:       cfg.Send,
			GasWanted:  root.GasWanted,
			GasFee:     root.GasFee,
			GasAutoEst: autoEst,
			GasEstUsed: gasEstUsed,
			Memo:       root.Memo,
			ChainID:    root.ChainID,
			Remote:     root.RootCfg.Remote,
		})
		proceed, err := promptProceed(io)
		if err != nil {
			return err
		}
		if !proceed {
			io.ErrPrintfln("Cancelled.")
			return nil
		}
	}

	if !root.Broadcast && canPrompt(root, io) {
		tx, err := buildSendTx(cfg, args[0])
		if err != nil {
			return err
		}
		return saveUnsignedTxAndPrintAirGap(tx, root, args[0], addr, io, "", "send")
	}

	return client.ExecMakeSend(cfg, args, io)
}

func buildSendTx(cfg *client.MakeSendCfg, nameOrBech32 string) (std.Tx, error) {
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.Home)
	if err != nil {
		return std.Tx{}, err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return std.Tx{}, err
	}
	toAddr, err := crypto.AddressFromBech32(cfg.To)
	if err != nil {
		return std.Tx{}, err
	}
	send, err := std.ParseCoins(cfg.Send)
	if err != nil {
		return std.Tx{}, errors.Wrap(err, "parsing send coins")
	}
	var gasfee std.Coin
	if cfg.RootCfg.GasFee != "" {
		gasfee, err = std.ParseCoin(cfg.RootCfg.GasFee)
		if err != nil {
			return std.Tx{}, errors.Wrap(err, "parsing gas fee coin")
		}
	}
	msg := bank.MsgSend{
		FromAddress: info.GetAddress(),
		ToAddress:   toAddr,
		Amount:      send,
	}
	return std.Tx{
		Msgs: []std.Msg{msg},
		Fee:  std.NewFee(cfg.RootCfg.GasWanted, gasfee),
		Memo: cfg.RootCfg.Memo,
	}, nil
}

func buildAddPkgTx(cfg *MakeAddPkgCfg, nameOrBech32 string) (std.Tx, error) {
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.Home)
	if err != nil {
		return std.Tx{}, err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return std.Tx{}, err
	}
	send, err := std.ParseCoins(cfg.Send)
	if err != nil {
		return std.Tx{}, errors.Wrap(err, "parsing send coins")
	}
	deposit, err := std.ParseCoins(cfg.MaxDeposit)
	if err != nil {
		return std.Tx{}, errors.Wrap(err, "parsing storage deposit coins")
	}
	memPkg := gno.MustReadMemPackage(cfg.PkgDir, cfg.PkgPath, gno.MPUserAll)
	if memPkg.IsEmpty() {
		return std.Tx{}, fmt.Errorf("empty package %q", cfg.PkgPath)
	}
	var gasfee std.Coin
	if cfg.RootCfg.GasFee != "" {
		gasfee, err = std.ParseCoin(cfg.RootCfg.GasFee)
		if err != nil {
			return std.Tx{}, errors.Wrap(err, "parsing gas fee coin")
		}
	}
	msg := vm.MsgAddPackage{
		Creator:    info.GetAddress(),
		Package:    memPkg,
		Send:       send,
		MaxDeposit: deposit,
	}
	return std.Tx{
		Msgs: []std.Msg{msg},
		Fee:  std.NewFee(cfg.RootCfg.GasWanted, gasfee),
		Memo: cfg.RootCfg.Memo,
	}, nil
}

func buildRunTx(cfg *MakeRunCfg, nameOrBech32, sourcePath string, cmdio commands.IO) (std.Tx, error) {
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.Home)
	if err != nil {
		return std.Tx{}, err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return std.Tx{}, err
	}
	send, err := std.ParseCoins(cfg.Send)
	if err != nil {
		return std.Tx{}, errors.Wrap(err, "parsing send coins")
	}
	deposit, err := std.ParseCoins(cfg.MaxDeposit)
	if err != nil {
		return std.Tx{}, errors.Wrap(err, "parsing storage deposit coins")
	}

	memPkg := &std.MemPackage{}
	if sourcePath == "-" {
		data, err := io.ReadAll(cmdio.In())
		if err != nil {
			return std.Tx{}, fmt.Errorf("could not read stdin: %w", err)
		}
		memPkg.Files = []*std.MemFile{{Name: "stdin.gno", Body: string(data)}}
	} else {
		info, err := os.Stat(sourcePath)
		if err != nil {
			return std.Tx{}, fmt.Errorf("could not read source path: %q, %w", sourcePath, err)
		}
		if info.IsDir() {
			memPkg = gno.MustReadMemPackage(sourcePath, "", gno.MPUserProd)
		} else {
			b, err := os.ReadFile(sourcePath)
			if err != nil {
				return std.Tx{}, fmt.Errorf("could not read %q: %w", sourcePath, err)
			}
			memPkg.Files = []*std.MemFile{{Name: info.Name(), Body: string(b)}}
		}
	}
	memPkg.Name = "main"
	if memPkg.IsEmpty() {
		return std.Tx{}, fmt.Errorf("empty package %q", memPkg.Path)
	}
	memPkg.Path = ""

	var gasfee std.Coin
	if cfg.RootCfg.GasFee != "" {
		gasfee, err = std.ParseCoin(cfg.RootCfg.GasFee)
		if err != nil {
			return std.Tx{}, errors.Wrap(err, "parsing gas fee coin")
		}
	}
	msg := vm.MsgRun{
		Caller:     info.GetAddress(),
		Package:    memPkg,
		Send:       send,
		MaxDeposit: deposit,
	}
	return std.Tx{
		Msgs: []std.Msg{msg},
		Fee:  std.NewFee(cfg.RootCfg.GasWanted, gasfee),
		Memo: cfg.RootCfg.Memo,
	}, nil
}

