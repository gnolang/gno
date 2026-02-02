package config

import (
	"errors"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/privval"
)

// -----------------------------------------------------------------------------
// ConsensusConfig

const (
	defaultWALDir = "wal"
)

// ConsensusConfig defines the configuration for the Tendermint consensus service,
// including timeouts and details about the WAL and the block structure.
type ConsensusConfig struct {
	RootDir     string `json:"home" toml:"home"`
	WALPath     string `json:"wal_file" toml:"wal_file"`
	WALDisabled bool   `json:"wal_disabled" toml:"-"`
	walFile     string // overrides WalPath if set

	PrivValidator *privval.PrivValidatorConfig `json:"priv_validator" toml:"priv_validator" comment:"##### private validator configuration options #####"`

	TimeoutPropose        time.Duration `json:"timeout_propose" toml:"timeout_propose"`
	TimeoutProposeDelta   time.Duration `json:"timeout_propose_delta" toml:"timeout_propose_delta"`
	TimeoutPrevote        time.Duration `json:"timeout_prevote" toml:"timeout_prevote"`
	TimeoutPrevoteDelta   time.Duration `json:"timeout_prevote_delta" toml:"timeout_prevote_delta"`
	TimeoutPrecommit      time.Duration `json:"timeout_precommit" toml:"timeout_precommit"`
	TimeoutPrecommitDelta time.Duration `json:"timeout_precommit_delta" toml:"timeout_precommit_delta"`
	TimeoutCommit         time.Duration `json:"timeout_commit" toml:"timeout_commit"`

	// Make progress as soon as we have all the precommits (as if TimeoutCommit = 0)
	SkipTimeoutCommit bool `json:"skip_timeout_commit" toml:"skip_timeout_commit" comment:"Make progress as soon as we have all the precommits (as if TimeoutCommit = 0)"`

	// EmptyBlocks mode and possible interval between empty blocks
	CreateEmptyBlocks         bool          `json:"create_empty_blocks" toml:"create_empty_blocks" comment:"EmptyBlocks mode and possible interval between empty blocks"`
	CreateEmptyBlocksInterval time.Duration `json:"create_empty_blocks_interval" toml:"create_empty_blocks_interval"`

	// Reactor sleep duration parameters
	PeerGossipSleepDuration     time.Duration `json:"peer_gossip_sleep_duration" toml:"peer_gossip_sleep_duration" comment:"Reactor sleep duration parameters"`
	PeerQueryMaj23SleepDuration time.Duration `json:"peer_query_maj_23_sleep_duration" toml:"peer_query_maj23_sleep_duration"`
}

// DefaultConsensusConfig returns a default configuration for the consensus service
func DefaultConsensusConfig() *ConsensusConfig {
	return &ConsensusConfig{
		WALPath:                     filepath.Join(defaultWALDir, "cs.wal", "wal"),
		PrivValidator:               privval.DefaultPrivValidatorConfig(),
		TimeoutPropose:              3000 * time.Millisecond,
		TimeoutProposeDelta:         500 * time.Millisecond,
		TimeoutPrevote:              1000 * time.Millisecond,
		TimeoutPrevoteDelta:         500 * time.Millisecond,
		TimeoutPrecommit:            1000 * time.Millisecond,
		TimeoutPrecommitDelta:       500 * time.Millisecond,
		TimeoutCommit:               5000 * time.Millisecond,
		SkipTimeoutCommit:           false,
		CreateEmptyBlocks:           true,
		CreateEmptyBlocksInterval:   0 * time.Second,
		PeerGossipSleepDuration:     100 * time.Millisecond,
		PeerQueryMaj23SleepDuration: 2000 * time.Millisecond,
	}
}

// TestConsensusConfig returns a configuration for testing the consensus service
func TestConsensusConfig() *ConsensusConfig {
	cfg := DefaultConsensusConfig()
	cfg.PrivValidator = privval.TestPrivValidatorConfig()
	cfg.TimeoutPropose = 500 * time.Millisecond
	cfg.TimeoutProposeDelta = 1 * time.Millisecond
	cfg.TimeoutPrevote = 100 * time.Millisecond
	cfg.TimeoutPrevoteDelta = 1 * time.Millisecond
	cfg.TimeoutPrecommit = 100 * time.Millisecond
	cfg.TimeoutPrecommitDelta = 1 * time.Millisecond
	cfg.TimeoutCommit = 100 * time.Millisecond
	cfg.SkipTimeoutCommit = true
	cfg.PeerGossipSleepDuration = 5 * time.Millisecond
	cfg.PeerQueryMaj23SleepDuration = 250 * time.Millisecond
	return cfg
}

// WaitForTxs returns true if the consensus should wait for transactions before entering the propose step
func (cfg *ConsensusConfig) WaitForTxs() bool {
	return !cfg.CreateEmptyBlocks || cfg.CreateEmptyBlocksInterval > 0
}

// Propose returns the amount of time to wait for a proposal
func (cfg *ConsensusConfig) Propose(round int) time.Duration {
	return time.Duration(
		cfg.TimeoutPropose.Nanoseconds()+cfg.TimeoutProposeDelta.Nanoseconds()*int64(round),
	) * time.Nanosecond
}

// Prevote returns the amount of time to wait for straggler votes after receiving any +2/3 prevotes
func (cfg *ConsensusConfig) Prevote(round int) time.Duration {
	return time.Duration(
		cfg.TimeoutPrevote.Nanoseconds()+cfg.TimeoutPrevoteDelta.Nanoseconds()*int64(round),
	) * time.Nanosecond
}

// Precommit returns the amount of time to wait for straggler votes after receiving any +2/3 precommits
func (cfg *ConsensusConfig) Precommit(round int) time.Duration {
	return time.Duration(
		cfg.TimeoutPrecommit.Nanoseconds()+cfg.TimeoutPrecommitDelta.Nanoseconds()*int64(round),
	) * time.Nanosecond
}

// Commit returns the amount of time to wait for straggler votes after receiving +2/3 precommits for a single block (ie. a commit).
func (cfg *ConsensusConfig) Commit(t time.Time) time.Time {
	return t.Add(cfg.TimeoutCommit)
}

// WalFile returns the full path to the write-ahead log file
func (cfg *ConsensusConfig) WalFile() string {
	if cfg.walFile != "" {
		return cfg.walFile
	}

	return filepath.Join(cfg.RootDir, cfg.WALPath)
}

// SetWalFile sets the path to the write-ahead log file
func (cfg *ConsensusConfig) SetWalFile(walFile string) {
	cfg.walFile = walFile
}

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *ConsensusConfig) ValidateBasic() error {
	if err := cfg.PrivValidator.ValidateBasic(); err != nil {
		return err
	}
	if cfg.TimeoutPropose < 0 {
		return errors.New("timeout_propose can't be negative")
	}
	if cfg.TimeoutProposeDelta < 0 {
		return errors.New("timeout_propose_delta can't be negative")
	}
	if cfg.TimeoutPrevote < 0 {
		return errors.New("timeout_prevote can't be negative")
	}
	if cfg.TimeoutPrevoteDelta < 0 {
		return errors.New("timeout_prevote_delta can't be negative")
	}
	if cfg.TimeoutPrecommit < 0 {
		return errors.New("timeout_precommit can't be negative")
	}
	if cfg.TimeoutPrecommitDelta < 0 {
		return errors.New("timeout_precommit_delta can't be negative")
	}
	if cfg.TimeoutCommit < 0 {
		return errors.New("timeout_commit can't be negative")
	}
	if cfg.CreateEmptyBlocksInterval < 0 {
		return errors.New("create_empty_blocks_interval can't be negative")
	}
	if cfg.PeerGossipSleepDuration < 0 {
		return errors.New("peer_gossip_sleep_duration can't be negative")
	}
	if cfg.PeerQueryMaj23SleepDuration < 0 {
		return errors.New("peer_query_maj23_sleep_duration can't be negative")
	}
	return nil
}
