package cluster

import "github.com/gnolang/gno/contribs/tessera/pkg/node"

// Config is the cluster configuration
type Config struct {
	Nodes []node.Config // the individual node configs
}
