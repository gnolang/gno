package integration

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/rogpeppe/go-internal/txtar"

	"github.com/gnolang/gno/misc/gnoe2e/internal/cluster"
)

// clusterSection is the txtar file a scenario declares its cluster in. A file
// section rather than a comment directive: the txtar comment block is the
// script testscript executes, so anything declared there would have to hide
// inside a "#" line.
const clusterSection = "cluster"

// approverRoleUser names the test user as the address allowed to enable inert
// packages, which is what leaves the oracle unauthorized.
const approverRoleUser = "user"

// The two generic key families, one per typed target they are applied to.
const (
	nodeConfigPrefix    = "config."
	genesisParamsPrefix = "genesis."
)

// ClusterSpec is the cluster one scenario declares it needs.
//
// It is the projection of cluster.ClusterConfig a scenario is allowed to
// choose, not a copy of it: the settings a script may state, in the vocabulary
// a script states them in, before the runner resolves them against the keys and
// binaries it owns.
type ClusterSpec struct {
	// Validators is the only setting with no usable zero value, so it is the
	// only one a scenario must state.
	Validators int
	// Oracle provisions the approver key and builds gpao. Off unless asked
	// for, so a scenario that never mentions the oracle does not pay to build
	// it.
	Oracle bool
	// CodeSubmissionPolicy is empty for the chain default.
	CodeSubmissionPolicy string
	// PkgApprover names who may enable inert packages, as a role rather than
	// an address, because the addresses are derived from mnemonics the runner
	// owns. Empty means the oracle. "user" means the test user, which is what
	// leaves the oracle unauthorized.
	PkgApprover string
	// BlockMaxGas is zero for the chain default.
	BlockMaxGas int64
	// NodeConfig holds the "config." keys, prefix stripped, in declaration
	// order. Each is a `gnoland config set` key set on every node.
	NodeConfig []cluster.Override
	// GenesisParams holds the "genesis." keys, prefix stripped, in declaration
	// order. Each is a `gnogenesis params set` key set on the genesis params.
	GenesisParams []cluster.Override
}

// ApplyTo settles the spec onto a cluster config, leaving every option the
// scenario did not state at the chain default.
//
// The approver is resolved here rather than in the section, because a script
// names a role while genesis needs an address, and the addresses come from
// mnemonics the runner owns.
func (s ClusterSpec) ApplyTo(cfg *cluster.ClusterConfig, user, oracle crypto.Address) error {
	cfg.NumValidators = s.Validators
	if s.CodeSubmissionPolicy != "" {
		cfg.Genesis.CodeSubmissionPolicy = vm.CodeSubmissionPolicy(s.CodeSubmissionPolicy)
	}
	if s.BlockMaxGas != 0 {
		cfg.Genesis.MaxGas = s.BlockMaxGas
	}
	cfg.NodeConfig = s.NodeConfig
	cfg.Genesis.Params = s.GenesisParams

	switch s.PkgApprover {
	case "":
		// An inert chain with no approver is one where nothing submitted after
		// genesis can ever go live, so it is refused rather than booted.
		if !s.Oracle && cfg.Genesis.CodeSubmissionPolicy == vm.CodeSubmissionPolicyInert {
			return fmt.Errorf("cluster section: an inert chain needs an approver: set %q or %q", "oracle: true", "pkg-approver: user")
		}
		if s.Oracle {
			cfg.Genesis.PkgApprovers = []crypto.Address{oracle}
		}
	case approverRoleUser:
		cfg.Genesis.PkgApprovers = []crypto.Address{user}
	default:
		// A command-line override names an address where a script names a
		// role, and both arrive here.
		addr, err := crypto.AddressFromBech32(s.PkgApprover)
		if err != nil {
			return fmt.Errorf("pkg-approver %q is neither %q nor an address: %w", s.PkgApprover, approverRoleUser, err)
		}
		cfg.Genesis.PkgApprovers = []crypto.Address{addr}
	}
	return nil
}

// ParseClusterSpec reads the cluster a script declares. Unknown keys are
// errors rather than ignored, so a typo fails at parse time instead of
// silently booting the wrong chain.
func ParseClusterSpec(script []byte) (ClusterSpec, error) {
	archive := txtar.Parse(script)

	var section []byte
	found := false
	for _, f := range archive.Files {
		if f.Name == clusterSection {
			section, found = f.Data, true
			break
		}
	}
	if !found {
		return ClusterSpec{}, fmt.Errorf("no %q section: every script must declare the cluster it needs", clusterSection)
	}

	spec, err := parseClusterSection(section)
	if err != nil {
		return ClusterSpec{}, fmt.Errorf("cluster section: %w", err)
	}
	return spec, nil
}

func parseClusterSection(section []byte) (ClusterSpec, error) {
	var spec ClusterSpec
	sawValidators := false
	for line := range strings.Lines(string(section)) {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return ClusterSpec{}, fmt.Errorf("expected %q, got %q", "key: value", line)
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)

		// The two generic families are handled ahead of the named keys: they
		// are a path rather than a key, so the switch below cannot match them.
		// A prefix with nothing after it falls through to the unknown-key
		// error, because it names no path at all.
		if path, ok := strings.CutPrefix(key, nodeConfigPrefix); ok && path != "" {
			// The listen addresses are the harness's: it picks a free port per
			// node and hands the addresses to the scripts, so a scenario
			// setting one would take the cluster away from its own commands.
			switch path {
			case "rpc.laddr", "p2p.laddr":
				return ClusterSpec{}, fmt.Errorf("%s cannot be set: the harness assigns each node its listen ports", key)
			}
			spec.NodeConfig = append(spec.NodeConfig, cluster.Override{Key: path, Value: value})
			continue
		}
		if path, ok := strings.CutPrefix(key, genesisParamsPrefix); ok && path != "" {
			spec.GenesisParams = append(spec.GenesisParams, cluster.Override{Key: path, Value: value})
			continue
		}

		switch key {
		case "validators":
			n, err := parseInt(key, value)
			if err != nil {
				return ClusterSpec{}, err
			}
			if n < 1 {
				return ClusterSpec{}, fmt.Errorf("validators must be at least 1, got %d", n)
			}
			spec.Validators, sawValidators = int(n), true
		case "block-max-gas":
			n, err := parseInt(key, value)
			if err != nil {
				return ClusterSpec{}, err
			}
			spec.BlockMaxGas = n
		case "oracle":
			b, err := strconv.ParseBool(value)
			if err != nil {
				return ClusterSpec{}, fmt.Errorf("%s %q: %w", key, value, unwrapSyntax(err))
			}
			spec.Oracle = b
		case "code-submission-policy":
			spec.CodeSubmissionPolicy = value
		case "pkg-approver":
			spec.PkgApprover = value
		default:
			return ClusterSpec{}, fmt.Errorf("unknown key %q", key)
		}
	}

	// Checked after the whole section so an unknown key is reported ahead of a
	// missing one: a typo'd "validatorz" is the likelier cause of both.
	if !sawValidators {
		return ClusterSpec{}, fmt.Errorf("%q is required", "validators")
	}
	return spec, nil
}

func parseInt(key, value string) (int64, error) {
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s %q: %w", key, value, unwrapSyntax(err))
	}
	return n, nil
}

// unwrapSyntax strips strconv's own repetition of the function name and input,
// which a caller has already put in the message.
func unwrapSyntax(err error) error {
	var numErr *strconv.NumError
	if errors.As(err, &numErr) {
		return numErr.Err
	}
	return err
}
