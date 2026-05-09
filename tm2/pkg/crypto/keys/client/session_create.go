package client

import (
	"context"
	"flag"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// expiresAtNoneKeyword is the sentinel value for --expires-at meaning "no expiry".
// We require an explicit choice rather than treating empty as no-expiry, so that
// device-login / agent sessions can't be accidentally created without time bound.
const expiresAtNoneKeyword = "none"

type SessionCreateCfg struct {
	RootCfg *SessionCfg

	PublicKey   string
	ExpiresAt   string
	AllowPaths  commands.StringArr
	SpendLimit  string
	SpendPeriod int64
}

// NewSessionCreateCmd creates a gnokey session create command
func NewSessionCreateCmd(rootCfg *SessionCfg, io commands.IO) *commands.Command {
	cfg := &SessionCreateCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "create",
			ShortUsage: "session create [flags] <master-key-name or address>",
			ShortHelp:  "create a session account",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSessionCreate(cfg, args, io)
		},
	)
}

func (c *SessionCreateCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.PublicKey,
		"pubkey",
		"",
		"the subaccount public key in bech32 format",
	)

	fs.StringVar(
		&c.ExpiresAt,
		"expires-at",
		"",
		"session expiry (REQUIRED): duration (24h, 7d, 4w; max ~4y), unix timestamp, or 'none' for no expiry",
	)

	fs.Var(
		&c.AllowPaths,
		"allow-paths",
		"per-msg restrictions (REQUIRED, repeatable). Use '*' for unrestricted, or list specific entries: vm/exec:gno.land/r/foo, vm/run, bank/send, bank/multisend",
	)

	fs.StringVar(
		&c.SpendLimit,
		"spend-limit",
		"",
		"max spend per period (optional; omitted = no spending)",
	)

	fs.Int64Var(
		&c.SpendPeriod,
		"spend-period",
		0,
		"seconds; 0 = lifetime cap",
	)
}

func execSessionCreate(cfg *SessionCreateCfg, args []string, io commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	if err := rejectSessionMasterFlag(cfg.RootCfg.RootCfg); err != nil {
		return err
	}
	if cfg.RootCfg.RootCfg.GasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}
	if cfg.RootCfg.RootCfg.GasFee == "" {
		return errors.New("gas-fee not specified")
	}
	if cfg.PublicKey == "" {
		return errors.New("pubkey must be specified")
	}
	if cfg.SpendPeriod < 0 {
		return errors.New("spend-period must be non-negative")
	}
	if len(cfg.AllowPaths) == 0 {
		return errors.New("--allow-paths is required (use '*' for unrestricted, or list entries like vm/exec:gno.land/r/foo, bank/send)")
	}
	for i, p := range cfg.AllowPaths {
		if err := validateAllowPathsEntryShape(p); err != nil {
			return fmt.Errorf("--allow-paths[%d] %q: %w", i, p, err)
		}
	}

	expiresAt, err := parseExpiresAt(cfg.ExpiresAt, time.Now())
	if err != nil {
		return errors.Wrap(err, "parsing expires-at")
	}

	// read account pubkey.
	nameOrBech32 := args[0]
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.RootCfg.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	masterAddr := info.GetAddress()

	// Parse the public key
	sessionPub, err := crypto.PubKeyFromBech32(cfg.PublicKey)
	if err != nil {
		return fmt.Errorf("unable to parse public key from bech32, %w", err)
	}

	// parse gas wanted & fee.
	gaswanted := cfg.RootCfg.RootCfg.GasWanted
	gasfee, err := std.ParseCoin(cfg.RootCfg.RootCfg.GasFee)
	if err != nil {
		return errors.Wrap(err, "parsing gas fee coin")
	}

	// construct msg & tx and marshal.
	msg := auth.MsgCreateSession{
		Creator:     masterAddr,
		SessionKey:  sessionPub,
		ExpiresAt:   expiresAt,
		AllowPaths:  cfg.AllowPaths,
		SpendPeriod: cfg.SpendPeriod,
	}
	if cfg.SpendLimit != "" {
		// Parse send amount.
		spendLimit, err := std.ParseCoins(cfg.SpendLimit)
		if err != nil {
			return errors.Wrap(err, "parsing spend limit coins")
		}
		msg.SpendLimit = spendLimit
	}

	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       cfg.RootCfg.RootCfg.Memo,
	}

	if cfg.RootCfg.RootCfg.Broadcast {
		err := ExecSignAndBroadcast(cfg.RootCfg.RootCfg, args, tx, io)
		if err != nil {
			return err
		}
	} else {
		io.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}

// parseExpiresAt converts an --expires-at argument into a unix timestamp.
//
// Accepted forms:
//   - "none": returns 0 (no expiry, valid until revoked)
//   - duration with suffix h/m/s/ms/us/ns/d/w: relative to now (e.g. 24h, 7d, 4w, 0.5d)
//   - bare integer: treated as a unix timestamp (must be in (now, now+MaxSessionDuration])
//
// Empty string returns an error to force explicit choice — device-login / agent
// sessions should never be created without a deliberate decision about expiry.
//
// Durations and bare timestamps are bounded by std.MaxSessionDuration (~4 years)
// to fail fast on the client rather than wasting gas on a tx the chain will
// reject in handleMsgCreateSession.
func parseExpiresAt(s string, now time.Time) (int64, error) {
	if s == "" {
		return 0, errors.New("--expires-at is required (e.g. 24h, 7d, or 'none' for no expiry)")
	}
	if s == expiresAtNoneKeyword {
		return 0, nil
	}

	maxSecs := int64(std.MaxSessionDuration)

	if secs, ok := parseDurationSeconds(s); ok {
		if secs <= 0 {
			return 0, fmt.Errorf("expires-at %q must be a positive duration", s)
		}
		if secs > maxSecs {
			return 0, fmt.Errorf("expires-at %q exceeds chain max of %dd", s, maxSecs/86400)
		}
		return now.Unix() + secs, nil
	}

	ts, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid expires-at %q: expected duration (24h, 7d), unix timestamp, or 'none'", s)
	}
	nowUnix := now.Unix()
	if ts <= nowUnix {
		return 0, fmt.Errorf("expires-at %q must be a future unix timestamp; for a duration, use a unit (e.g. 24h, 7d)", s)
	}
	if ts-nowUnix > maxSecs {
		return 0, fmt.Errorf("expires-at %q exceeds chain max of %dd from now", s, maxSecs/86400)
	}
	return ts, nil
}

// parseDurationSeconds returns the input expressed as int64 seconds. Extends
// time.ParseDuration with fractional 'd' (days) and 'w' (weeks) suffixes.
// Overflow / NaN / Inf are clamped to the int64 limits so callers reject via
// upper-bound checks rather than computing a wrapped value.
// Returns (0, false) if the input isn't a recognizable duration.
func parseDurationSeconds(s string) (int64, bool) {
	var unitSecs float64
	var num string
	switch {
	case strings.HasSuffix(s, "w"):
		unitSecs, num = 7*86400, strings.TrimSuffix(s, "w")
	case strings.HasSuffix(s, "d"):
		unitSecs, num = 86400, strings.TrimSuffix(s, "d")
	default:
		dur, err := time.ParseDuration(s)
		if err != nil {
			return 0, false
		}
		return int64(dur / time.Second), true
	}
	n, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0, false
	}
	secs := n * unitSecs
	if math.IsNaN(secs) || secs > float64(math.MaxInt64) {
		return math.MaxInt64, true
	}
	if math.IsInf(secs, -1) || secs < float64(math.MinInt64) {
		return math.MinInt64, true
	}
	return int64(secs), true
}

// validateAllowPathsEntryShape is a CLI-side shape check for one
// --allow-paths entry. It enforces only the structural grammar
// ("*" or <route>/<type>[:<path>]); the chain enforces which
// route_types are real (e.g. "vm/foo" passes here but is rejected
// at create-time).
//
// Catches common typos at the CLI: bare "bank" (missing /), trailing
// slash, empty path after ':', extra slashes in the route_type, and
// "*" with a path suffix.
func validateAllowPathsEntryShape(s string) error {
	if s == "" {
		return errors.New("entry is empty")
	}
	if s == "*" {
		return nil
	}
	if strings.HasSuffix(s, "/") {
		return errors.New("entry must not end with /")
	}
	routeType, path, hasPath := strings.Cut(s, ":")
	if routeType == "*" {
		return errors.New("wildcard '*' must not have a path suffix")
	}
	parts := strings.Split(routeType, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return errors.New("expected '*' or <route>/<type>[:<path>] (e.g. vm/exec:gno.land/r/foo, bank/send)")
	}
	if hasPath && path == "" {
		return errors.New("path after ':' must be non-empty")
	}
	return nil
}
