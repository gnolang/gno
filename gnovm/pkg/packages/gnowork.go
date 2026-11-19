package packages

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml"
)

type GnoworkDomain struct {
	RPC string
}

type Gnowork struct {
	Domains map[string]GnoworkDomain
}

func (gw *Gnowork) rpcOverrides() map[string]string {
	if gw == nil {
		return nil
	}
	res := map[string]string{}
	for domainName, domain := range gw.Domains {
		if domain.RPC == "" {
			continue
		}
		res[domainName] = domain.RPC
	}
	return res
}

func ParseGnowork(bz []byte) (*Gnowork, error) {
	var res Gnowork
	if err := toml.Unmarshal(bz, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func ReadGnowork(file string) (*Gnowork, error) {
	bz, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read gnowork file %q: %w", file, err)
	}
	gw, err := ParseGnowork(bz)
	if err != nil {
		return nil, fmt.Errorf("parse gnowork file %q: %w", file, err)
	}
	return gw, nil
}
