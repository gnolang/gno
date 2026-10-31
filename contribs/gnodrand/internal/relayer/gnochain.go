package relayer

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// GnoChain talks to r/drand through gnoclient.
//
// Gas and fee are set per beacon and scaled by the batch size. The chain
// deducts the whole fee whatever the tx uses, so FeePerBeacon should be
// GasPerBeacon times the gas price, not a flat round number.
type GnoChain struct {
	Client       *gnoclient.Client
	PkgPath      string   // e.g. gno.land/r/drand/v0
	FeePerBeacon std.Coin // e.g. 22000ugnot
	GasPerBeacon int64    // e.g. 22000000
}

func (c GnoChain) PendingRounds(_ context.Context, limit int) ([]uint64, error) {
	res, _, err := c.Client.QEval(c.PkgPath, fmt.Sprintf("PendingRounds(%d)", limit))
	if err != nil {
		return nil, err
	}
	return parsePendingRounds(res)
}

// Submit sends one transaction with a Submit call per beacon.
func (c GnoChain) Submit(_ context.Context, beacons []Beacon) error {
	info, err := c.Client.Signer.Info()
	if err != nil {
		return err
	}
	msgs := make([]vm.MsgCall, 0, len(beacons))
	for _, b := range beacons {
		msgs = append(msgs, vm.MsgCall{
			Caller:  info.GetAddress(),
			PkgPath: c.PkgPath,
			Func:    "Submit",
			Args:    []string{strconv.FormatUint(b.Round, 10), b.Signature},
		})
	}
	n := int64(len(beacons))
	fee := std.Coin{Denom: c.FeePerBeacon.Denom, Amount: c.FeePerBeacon.Amount * n}
	_, err = c.Client.Call(gnoclient.BaseTxCfg{
		GasFee:    fee.String(),
		GasWanted: c.GasPerBeacon * n,
		Memo:      "gnodrand relayer",
	}, msgs...)
	return err
}

// parsePendingRounds decodes the qeval output of PendingRounds, which looks
// like `("12,15,20" string)`.
func parsePendingRounds(res string) ([]uint64, error) {
	res = strings.TrimSpace(res)
	const prefix, suffix = `("`, `" string)`
	if !strings.HasPrefix(res, prefix) || !strings.HasSuffix(res, suffix) {
		return nil, fmt.Errorf("unexpected qeval output %q", res)
	}
	body := strings.TrimSuffix(strings.TrimPrefix(res, prefix), suffix)
	if body == "" {
		return nil, nil
	}
	parts := strings.Split(body, ",")
	rounds := make([]uint64, 0, len(parts))
	for _, p := range parts {
		r, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad round %q: %w", p, err)
		}
		rounds = append(rounds, r)
	}
	return rounds, nil
}
