package common

import (
	"flag"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

const DefaultChainID = "dev"

// Cfg is the common
// configuration for genesis commands
// that require a genesis.json
type Cfg struct {
	GenesisPath string
}

func (c *Cfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.GenesisPath,
		"genesis-path",
		"./genesis.json",
		"the path to the genesis.json",
	)
}

// DefaultGenesis returns the default genesis config
func DefaultGenesis() *types.GenesisDoc {
	return &types.GenesisDoc{
		GenesisTime:     time.Now(),
		ChainID:         DefaultChainID,
		ConsensusParams: types.DefaultConsensusParams(),
		AppState:        gnoland.DefaultGenState(),
	}
}
