package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type configConsensusCfg struct {
	commonEditCfg

	rootDir                     string
	walPath                     string
	timeoutPropose              time.Duration
	timeoutProposeDelta         time.Duration
	timeoutPrevote              time.Duration
	timeoutPrevoteDelta         time.Duration
	timeoutPrecommit            time.Duration
	timeoutPrecommitDelta       time.Duration
	timeoutCommit               time.Duration
	skipTimeoutCommit           string // toggle
	createEmptyBlocks           string // toggle
	createEmptyBlocksInterval   time.Duration
	peerGossipSleepDuration     time.Duration
	peerQueryMaj23SleepDuration time.Duration
}

// newConfigConsensusCmd creates the new config consensus command
func newConfigConsensusCmd(io commands.IO) *commands.Command {
	cfg := &configConsensusCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "consensus",
			ShortUsage: "config consensus [flags]",
			ShortHelp:  "Edits the Gno node's consensus configuration",
			LongHelp:   "Edits the Gno node's consensus configuration locally",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execConfigConsensus(cfg, io)
		},
	)

	return cmd
}

func (c *configConsensusCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonEditCfg.RegisterFlags(fs)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"the root directory for all consensus data",
	)

	fs.StringVar(
		&c.walPath,
		"wal-path",
		"",
		"the path to the WAL file",
	)

	fs.DurationVar(
		&c.timeoutPropose,
		"timeout-propose",
		time.Second*0,
		"the propose phase timeout",
	)

	fs.DurationVar(
		&c.timeoutProposeDelta,
		"timeout-propose-delta",
		time.Second*0,
		"the propose phase timeout delta",
	)

	fs.DurationVar(
		&c.timeoutPrevote,
		"timeout-prevote",
		time.Second*0,
		"the prevote phase timeout",
	)

	fs.DurationVar(
		&c.timeoutPrevoteDelta,
		"timeout-prevote-delta",
		time.Second*0,
		"the prevote phase timeout delta",
	)

	fs.DurationVar(
		&c.timeoutPrecommit,
		"timeout-precommit",
		time.Second*0,
		"the precommit phase timeout",
	)

	fs.DurationVar(
		&c.timeoutPrecommitDelta,
		"timeout-precommit-delta",
		time.Second*0,
		"the precommit phase timeout delta",
	)

	fs.DurationVar(
		&c.timeoutCommit,
		"timeout-commit",
		time.Second*0,
		"the commit phase timeout",
	)

	fs.StringVar(
		&c.skipTimeoutCommit,
		"skip-commit-timeout",
		offValue,
		fmt.Sprintf(
			"toggle value indicating if progress"+
				" should be made as soon as we have all the precommits (as if TimeoutCommit = 0): %s | %s",
			onValue,
			offValue,
		),
	)

	fs.StringVar(
		&c.createEmptyBlocks,
		"create-empty-blocks",
		onValue,
		fmt.Sprintf(
			"toggle value indicating if empty blocks should be made: %s | %s",
			onValue,
			offValue,
		),
	)

	fs.DurationVar(
		&c.createEmptyBlocksInterval,
		"empty-blocks-interval",
		time.Second*0,
		"interval for creating empty blocks",
	)

	fs.DurationVar(
		&c.peerGossipSleepDuration,
		"gossip-sleep-duration",
		time.Second*0,
		"the peer gossip sleep duration",
	)

	fs.DurationVar(
		&c.peerQueryMaj23SleepDuration,
		"query-sleep-duration",
		time.Second*0,
		"the peer query majority sleep duration",
	)
}

func execConfigConsensus(cfg *configConsensusCfg, io commands.IO) error {
	// Load the config
	loadedCfg, err := config.LoadConfigFile(cfg.configPath)
	if err != nil {
		return fmt.Errorf("unable to load config, %w", err)
	}

	// Set the root dir, if any
	if cfg.rootDir != "" {
		loadedCfg.Consensus.RootDir = cfg.rootDir
	}

	// Set the WAL path, if any
	if cfg.walPath != "" {
		loadedCfg.Consensus.WalPath = cfg.walPath
	}

	// Set the propose timeout, if any
	if cfg.timeoutPropose != time.Second*0 {
		loadedCfg.Consensus.TimeoutPropose = cfg.timeoutPropose
	}

	// Set the propose timeout delta, if any
	if cfg.timeoutProposeDelta != time.Second*0 {
		loadedCfg.Consensus.TimeoutProposeDelta = cfg.timeoutProposeDelta
	}

	// Set the prevote timeout, if any
	if cfg.timeoutPrevote != time.Second*0 {
		loadedCfg.Consensus.TimeoutPrevote = cfg.timeoutPrevote
	}

	// Set the prevote timeout delta, if any
	if cfg.timeoutPrevoteDelta != time.Second*0 {
		loadedCfg.Consensus.TimeoutPrevoteDelta = cfg.timeoutPrevoteDelta
	}

	// Set the precommit timeout, if any
	if cfg.timeoutPrecommit != time.Second*0 {
		loadedCfg.Consensus.TimeoutPrecommit = cfg.timeoutPrecommit
	}

	// Set the precommit timeout delta, if any
	if cfg.timeoutPrecommitDelta != time.Second*0 {
		loadedCfg.Consensus.TimeoutPrecommitDelta = cfg.timeoutPrecommitDelta
	}

	// Set the commit timeout, if any
	if cfg.timeoutCommit != time.Second*0 {
		loadedCfg.Consensus.TimeoutCommit = cfg.timeoutCommit
	}

	// Set the skip commit timeout toggle, if any
	skipTimeoutCommitVal, err := parseToggleValue(cfg.skipTimeoutCommit)
	if err != nil {
		return err
	}

	if skipTimeoutCommitVal != loadedCfg.Consensus.SkipTimeoutCommit {
		loadedCfg.Consensus.SkipTimeoutCommit = skipTimeoutCommitVal
	}

	// Set the skip commit timeout toggle, if any
	createEmptyBlocksVal, err := parseToggleValue(cfg.createEmptyBlocks)
	if err != nil {
		return err
	}

	if createEmptyBlocksVal != loadedCfg.Consensus.CreateEmptyBlocks {
		loadedCfg.Consensus.CreateEmptyBlocks = createEmptyBlocksVal
	}

	// Set the create empty blocks interval, if any
	if cfg.createEmptyBlocksInterval != time.Second*0 {
		loadedCfg.Consensus.CreateEmptyBlocksInterval = cfg.createEmptyBlocksInterval
	}

	// Set the peer gossip sleep duration, if any
	if cfg.peerGossipSleepDuration != time.Second*0 {
		loadedCfg.Consensus.PeerGossipSleepDuration = cfg.peerGossipSleepDuration
	}

	// Set the peer majority query sleep duration, if any
	if cfg.peerQueryMaj23SleepDuration != time.Second*0 {
		loadedCfg.Consensus.PeerQueryMaj23SleepDuration = cfg.peerQueryMaj23SleepDuration
	}

	// Make sure the config is now valid
	if err := loadedCfg.ValidateBasic(); err != nil {
		return fmt.Errorf("unable to validate config, %w", err)
	}

	// Save the config
	if err := config.WriteConfigFile(cfg.configPath, loadedCfg); err != nil {
		return fmt.Errorf("unable to save updated config, %w", err)
	}

	io.Printfln("Updated consensus configuration saved at %s", cfg.configPath)

	return nil
}
