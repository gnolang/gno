package main

import (
	"testing"

	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/misc/gnoe2e/internal/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serve boots straight from its flags, with no scenario to fill anything in, so
// the chain they describe is the chain that runs. A genesis the node would
// refuse has to be refused here rather than after the ~100 MB gnoland build the
// command would otherwise do first.
func TestServeRefusesAChainNothingCanEverGoLiveOn(t *testing.T) {
	tests := map[string]func(cfg *cluster.ClusterConfig){
		"the named key, with nobody named to enable": func(cfg *cluster.ClusterConfig) {
			cfg.Genesis.CodeSubmissionPolicy = vm.CodeSubmissionPolicyInert
		},
		"the path spelling, which never touches the named field": func(cfg *cluster.ClusterConfig) {
			cfg.Genesis.Params = []cluster.Override{{Key: "vm.code_submission_policy", Value: "inert"}}
		},
	}

	for name, declare := range tests {
		t.Run(name, func(t *testing.T) {
			clusterCfg := cluster.DefaultClusterConfig()
			declare(&clusterCfg)

			err := (&serveCfg{ClusterConfig: clusterCfg}).Validate()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "approver")
		})
	}
}

// A cluster somebody can enable on is served, whichever spelling named them.
func TestServeAcceptsAnInertChainWithAnApprover(t *testing.T) {
	clusterCfg := cluster.DefaultClusterConfig()
	clusterCfg.Genesis.Params = []cluster.Override{
		{Key: "vm.code_submission_policy", Value: "inert"},
		{Key: "vm.pkg_approvers", Value: "g19rl4cm2hmr8afy4kldpxz3fka4jguq0a0u3773"},
	}

	require.NoError(t, (&serveCfg{ClusterConfig: clusterCfg}).Validate())
}
