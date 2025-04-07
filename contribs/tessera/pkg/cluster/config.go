package cluster

import (
	"time"

	"github.com/gnolang/gno/contribs/tessera/pkg/node"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

// Config is the cluster configuration
type Config struct {
	Genesis GenesisCfg    `yaml:"genesis"` // variable genesis params
	Nodes   []node.Config `yaml:"nodes"`   // the individual node configs
}

// GenesisCfg allows for configuration of specific genesis params.
// Validator and AppState management are not exposed, but managed by
// the tool. This might not always be the case
type GenesisCfg struct {
	GenesisTime     time.Time            `yaml:"genesis_time"`
	ChainID         string               `yaml:"chain_id"`
	ConsensusParams abci.ConsensusParams `yaml:"consensus_params,omitempty"`
}
