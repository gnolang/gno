#!/bin/bash
# Recalibrate GnoVM CPU gas constants after optimization changes.
# Run on the reference hardware (Xeon 8168) for consistent results.
#
# Usage:
#   cd gnovm/cmd/calibrate
#   bash recalibrate.sh
#
# Prerequisites:
#   go install golang.org/x/perf/cmd/benchstat@latest

set -euo pipefail

REPO_ROOT="$(cd ../../.. && pwd)"
cd "$REPO_ROOT"

BRANCH=$(git branch --show-current)
BASE_COMMIT="ae5c8ae2d"  # benchmark commit (before optimizations)
OUTDIR="/tmp/gas_calibrate"
mkdir -p "$OUTDIR"

echo "=== Gas Constant Recalibration ==="
echo "Branch: $BRANCH"
echo "Base:   $BASE_COMMIT"
echo "Output: $OUTDIR"
echo

# Benchmarks to run — covers all op categories affected by optimizations.
# Note: BenchmarkOpMethodPrecall_ValMethod only exists in the optimized code.
# It won't run on the base commit (that's expected — it's a new op).
BENCH_PATTERN='BenchmarkOpAdd_Int$|BenchmarkOpAdd_String_100$|BenchmarkOpAdd_Float64$|BenchmarkOpSub_Int$|BenchmarkOpMul_Int$|BenchmarkOpQuo_Int$|BenchmarkOpRem_Int$|BenchmarkOpCall_0Params_0Captures$|BenchmarkOpCall_1Params_0Captures$|BenchmarkOpCall_10Params_0Captures$|BenchmarkOpCall_0Params_1Captures$|BenchmarkOpCall_0Params_10Captures$|BenchmarkOpCall_10Params_10Captures$|BenchmarkOpCall_Method$|BenchmarkOpPrecall_BoundMethod$|BenchmarkOpMethodPrecall_ValMethod$|BenchmarkOpSelector_VPValMethod$|BenchmarkOpSelector_VPInterface_1$|BenchmarkOpSelector_VPInterface_10$|BenchmarkOpStructLit_Unnamed_3$|BenchmarkOpStructLit_Named_3$|BenchmarkOpReturn$|BenchmarkOpForLoop$|BenchmarkOpIfCond$|BenchmarkOpEval_NameExpr_Depth1$|BenchmarkOpEval_NameExpr_Depth10$'

BENCH_FLAGS="-benchtime=3s -count=5"

echo "--- Step 1: Benchmark AFTER (optimized) ---"
go test ./gnovm/pkg/gnolang/ -bench="$BENCH_PATTERN" $BENCH_FLAGS \
  2>&1 | tee "$OUTDIR/ops_after.txt"

echo
echo "--- Step 2: Benchmark BEFORE (pre-optimization) ---"
git stash --quiet 2>/dev/null || true
git checkout "$BASE_COMMIT" --quiet
go test ./gnovm/pkg/gnolang/ -bench="$BENCH_PATTERN" $BENCH_FLAGS \
  2>&1 | tee "$OUTDIR/ops_before.txt"

echo
echo "--- Step 3: Restore branch ---"
git checkout "$BRANCH" --quiet
git stash pop --quiet 2>/dev/null || true

echo
echo "--- Step 4: Compare ---"
echo
benchstat "$OUTDIR/ops_before.txt" "$OUTDIR/ops_after.txt"

echo
echo "=== Interpretation ==="
echo "If all ops changed by a similar percentage (uniform speedup),"
echo "only cpuBaseNs needs updating — the gas constants stay the same."
echo
echo "If specific ops changed disproportionately (e.g., OpCall got 40%"
echo "faster but OpAdd stayed the same), those ops' gas constants need"
echo "adjustment to maintain proportional charging."
echo
echo "Key ops to watch:"
echo "  OpCall              — block arena makes this cheaper"
echo "  OpCall_Method       — no BoundMethodValue allocation"
echo "  OpMethodPrecall     — NEW op replacing OpSelector+OpPrecall for methods"
echo "  OpSelector_VPInterface — cached method trail"
echo "  OpQuo/OpRem         — lazy exception allocation"
echo "  OpAdd_Int           — baseline reference (should not change)"
echo
echo "New op: OpMethodPrecall replaces OpSelector+OpPrecall for direct method"
echo "calls. Its gas cost should be roughly OpCPUSelector + OpCPUPrecall since"
echo "it does both in one step (but cheaper due to no heap allocation)."
echo "The old OpPrecall_BoundMethod path is still used for stored method values."
echo
echo "Results saved to: $OUTDIR/"
