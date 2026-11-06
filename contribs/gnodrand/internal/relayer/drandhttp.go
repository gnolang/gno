package relayer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gnolang/gno/contribs/gnodrand/internal/evmnet"
)

// DefaultMirrors are the public drand HTTP endpoints, tried in order.
var DefaultMirrors = []string{
	"https://api.drand.sh",
	"https://api2.drand.sh",
	"https://api3.drand.sh",
	"https://drand.cloudflare.com",
}

// HTTPSource fetches evmnet beacons from drand HTTP mirrors, falling back to
// the next mirror on any error.
type HTTPSource struct {
	Mirrors []string
	Client  *http.Client
}

func (s HTTPSource) Beacon(ctx context.Context, round uint64) (Beacon, error) {
	var errs []error
	for _, m := range s.Mirrors {
		b, err := s.fetch(ctx, m, round)
		if err == nil {
			return b, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", m, err))
	}
	return Beacon{}, errors.Join(errs...)
}

func (s HTTPSource) fetch(ctx context.Context, mirror string, round uint64) (Beacon, error) {
	url := fmt.Sprintf("%s/%s/public/%d", strings.TrimRight(mirror, "/"), evmnet.ChainHash, round)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Beacon{}, err
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return Beacon{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return Beacon{}, fmt.Errorf("http %d", resp.StatusCode)
	}
	var b Beacon
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return Beacon{}, fmt.Errorf("decode: %w", err)
	}
	return b, nil
}
