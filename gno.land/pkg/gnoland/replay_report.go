package gnoland

import (
	"fmt"
	"log/slog"

	"github.com/gnolang/gno/tm2/pkg/sdk"
)

// ReplayCategory classifies the outcome of a genesis tx replay.
type ReplayCategory string

const (
	// ReplayCategoryOK: tx replayed successfully (gas matched source within tolerance, if source gas was recorded).
	ReplayCategoryOK ReplayCategory = "ok"
	// ReplayCategoryOKGasDiffers: tx succeeded but gas consumption differs from source chain.
	ReplayCategoryOKGasDiffers ReplayCategory = "ok_gas_differs"
	// ReplayCategoryFailed: tx failed during replay (any reason not covered by specific categories).
	ReplayCategoryFailed ReplayCategory = "failed"
	// ReplayCategorySkippedFailed: tx was marked Failed in source metadata, correctly skipped.
	ReplayCategorySkippedFailed ReplayCategory = "skipped_failed"

	// aliases for callers (lowercase internal):
	replayCategorySkippedFailed = ReplayCategorySkippedFailed
)

// replayOutcome is a single tx outcome during genesis replay.
type replayOutcome struct {
	TxIndex       int            `json:"tx_index"`
	SourceHeight  int64          `json:"source_height,omitempty"` // metadata.BlockHeight
	SourceChainID string         `json:"source_chain_id,omitempty"`
	Category      ReplayCategory `json:"category"`
	GasSource     int64          `json:"gas_source,omitempty"` // metadata.GasUsed (from tx-archive)
	GasReplay     int64          `json:"gas_replay,omitempty"` // actual gas consumed during replay
	Error         string         `json:"error,omitempty"`      // brief error if failed
}

// replayReport accumulates per-tx outcomes and emits a summary.
type replayReport struct {
	mode     string // GasReplayMode from GnoGenesisState
	outcomes []replayOutcome
}

func newReplayReport(mode string) *replayReport {
	return &replayReport{mode: mode}
}

// record appends an outcome with fully explicit values (used for skipped txs).
func (r *replayReport) record(txIdx int, metadata *GnoTxMetadata, gasReplay int64, gasSource int64, cat ReplayCategory, err error) {
	o := replayOutcome{
		TxIndex:   txIdx,
		Category:  cat,
		GasReplay: gasReplay,
		GasSource: gasSource,
	}
	if metadata != nil {
		o.SourceHeight = metadata.BlockHeight
		o.SourceChainID = metadata.ChainID
		if o.GasSource == 0 {
			o.GasSource = metadata.GasUsed
		}
	}
	if err != nil {
		o.Error = err.Error()
	}
	r.outcomes = append(r.outcomes, o)
}

// recordDeliverResult derives the outcome from a Deliver result and metadata.
func (r *replayReport) recordDeliverResult(txIdx int, metadata *GnoTxMetadata, res sdk.Result) {
	o := replayOutcome{
		TxIndex:   txIdx,
		GasReplay: res.GasUsed,
	}
	if metadata != nil {
		o.SourceHeight = metadata.BlockHeight
		o.SourceChainID = metadata.ChainID
		o.GasSource = metadata.GasUsed
	}
	if res.IsErr() {
		o.Category = ReplayCategoryFailed
		if res.Error != nil {
			o.Error = res.Error.Error()
		} else if res.Log != "" {
			o.Error = res.Log
		}
	} else if o.GasSource > 0 && o.GasReplay != o.GasSource {
		o.Category = ReplayCategoryOKGasDiffers
	} else {
		o.Category = ReplayCategoryOK
	}
	r.outcomes = append(r.outcomes, o)
}

// emit writes a summary to the logger.
func (r *replayReport) emit(logger *slog.Logger) {
	if logger == nil || len(r.outcomes) == 0 {
		return
	}
	counts := map[ReplayCategory]int{}
	for _, o := range r.outcomes {
		counts[o.Category]++
	}
	logger.Info(
		"Genesis replay report",
		"mode", modeOrDefault(r.mode),
		"total", len(r.outcomes),
		"ok", counts[ReplayCategoryOK],
		"ok_gas_differs", counts[ReplayCategoryOKGasDiffers],
		"failed", counts[ReplayCategoryFailed],
		"skipped_failed", counts[ReplayCategorySkippedFailed],
	)
	// For failures, log each one so operators can review.
	for _, o := range r.outcomes {
		if o.Category == ReplayCategoryFailed {
			logger.Warn("Genesis replay failure",
				"tx_index", o.TxIndex,
				"source_height", o.SourceHeight,
				"gas_source", o.GasSource,
				"gas_replay", o.GasReplay,
				"error", o.Error,
			)
		}
	}
}

// Outcomes returns a copy of recorded outcomes. Exposed for tests and tooling
// that wants to write its own replay-report.json.
func (r *replayReport) Outcomes() []replayOutcome {
	out := make([]replayOutcome, len(r.outcomes))
	copy(out, r.outcomes)
	return out
}

// FailedCount returns the number of outcomes categorized as ReplayCategoryFailed
// (real replay failures), excluding ReplayCategorySkippedFailed (txs intentionally
// skipped because they were marked Failed in source metadata).
func (r *replayReport) FailedCount() int {
	n := 0
	for _, o := range r.outcomes {
		if o.Category == ReplayCategoryFailed {
			n++
		}
	}
	return n
}

func modeOrDefault(mode string) string {
	if mode == "" {
		return "strict"
	}
	return mode
}

// validateGasReplayMode returns an error if the given mode is not recognised.
func validateGasReplayMode(mode string) error {
	switch mode {
	case "", "strict", "source":
		return nil
	default:
		return fmt.Errorf("unknown GasReplayMode %q (valid: \"\", \"strict\", \"source\")", mode)
	}
}
