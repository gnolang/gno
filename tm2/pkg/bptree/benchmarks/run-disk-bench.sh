#!/usr/bin/env bash
#
# Disk-bound IAVL vs B+32 benchmark driver.
#
# Populates a per-factory fixture once (resumable), then benchmarks block writes
# and random reads against it. See BENCHMARKS.md for what the numbers mean and
# which scale is meaningful: N must exceed RAM to be disk-bound (>~57M on 16 GB;
# below that everything is cached and only the per-op *counts* are informative).
#
# Run from anywhere in the repo:
#   DIR=/data/bp32bench KEYS=100000000 ./tm2/pkg/bptree/benchmarks/run-disk-bench.sh
#
# Override anything via env vars (defaults in brackets):
#   DIR         fixture dir, needs ~3x per-tree size free   [/data/bp32bench]
#   KEYS        fixture size; >57M on 16 GB to be disk-bound [100000000]
#   BACKEND     pebbledb | lmdbdb | goleveldb | boltdb       [pebbledb]
#               (reads/writes counts are backend-agnostic; ns is not.
#                lmdbdb is the backend gno.land's flat gas costs reference.)
#   BLOCK       writes per block (SaveVersion cadence)        [1000]
#   WRITE_N     BlockWrite benchtime, in blocks              [1200]
#               (16 convergence windows of 75 blocks; lower for quick checks)
#   READ_N      GetRandom/GetMiss benchtime, in ops        [4000000]
#               (40 reload-interval windows; capped at 8 reported.
#                lower for quick checks)
#   NODE_CACHE  in-process node LRU                           [10000]
#   FACTORIES   trees to run                                  ["iavl bptree-fast"]
#               (bptree-fast reuses the "bptree" fixture and backfills the
#                inline index on first Load — see the heads-up below)
#   BUILD_BATCH keys per SaveVersion while building           [25000]
#               (raise to 100000 on a big-RAM box for a faster build)
#   WARMUP      untimed ops before each bench                 [50000]
#               (warms the node LRU; reported counts average over every
#               iteration, so a cold start inflates them. Size to several
#               x NODE_CACHE: 50000 fits the default 10K cache; use
#               ~1500000 when benchmarking a 330K-node cache.)
#   PARALLEL    1 = populate both factories at once           [1]
#               (faster, but doubles RAM pressure; set 0 on a tight box)
#   DROP_CACHES 1 = drop OS page cache before each bench       [1] (needs sudo)
#   GOMEMLIMIT  Go soft heap cap, to survive big populates     [12GiB]
#   OUT         results/logs dir                               [./bench-out]
#   COMMITTED_READ 1 = GetRandom/GetMiss read via a committed    [1]
#               snapshot at the latest version (the ABCI-query path the bptree
#               fast index serves); 0 = working-tree Get (index-free, the old
#               behavior). bptree-fast only beats iavl on reads with this on.

set -uo pipefail
cd "$(git rev-parse --show-toplevel)"

DIR="${DIR:-/data/bp32bench}"
KEYS="${KEYS:-100000000}"
BACKEND="${BACKEND:-pebbledb}"
BLOCK="${BLOCK:-1000}"
WRITE_N="${WRITE_N:-1200}"
READ_N="${READ_N:-4000000}"
NODE_CACHE="${NODE_CACHE:-10000}"
FACTORIES="${FACTORIES:-iavl bptree-fast}"
BUILD_BATCH="${BUILD_BATCH:-25000}"
WARMUP="${WARMUP:-50000}"
PARALLEL="${PARALLEL:-1}"
DROP_CACHES="${DROP_CACHES:-1}"
OUT="${OUT:-./bench-out}"
COMMITTED_READ="${COMMITTED_READ:-1}"
export GOMEMLIMIT="${GOMEMLIMIT:-12GiB}"

PKG=./tm2/pkg/bptree/benchmarks/
mkdir -p "$DIR" "$OUT"
common=(-disk-dir="$DIR" -disk-keys="$KEYS" -disk-backend="$BACKEND" -disk-node-cache="$NODE_CACHE")
# committed-read flag for the Get benches only (BlockWrite ignores it). A scalar,
# not an array, so the empty case expands to nothing cleanly under `set -u`.
read_flag=""
[ "$COMMITTED_READ" = 1 ] && read_flag="-disk-committed-read"

drop_caches() {
	[ "$DROP_CACHES" = 1 ] || return 0
	sync
	if [ -w /proc/sys/vm/drop_caches ]; then
		echo 3 >/proc/sys/vm/drop_caches
	elif [ -e /proc/sys/vm/drop_caches ] && command -v sudo >/dev/null; then
		echo 3 | sudo tee /proc/sys/vm/drop_caches >/dev/null
	elif command -v purge >/dev/null; then
		sudo purge 2>/dev/null || purge 2>/dev/null || true # darwin
	fi
}

# 1. populate (resumable: re-running continues from the current size).
populate() {
	local f=$1
	echo ">> populate $f -> $KEYS keys ($BACKEND)  [log: $OUT/populate-$f.log]"
	if go test "$PKG" -run=TestDiskPopulate -v -timeout=24h \
		"${common[@]}" -disk-factory="$f" -disk-build-batch="$BUILD_BATCH" -disk-verbose \
		>"$OUT/populate-$f.log" 2>&1; then
		echo ">> populate $f OK"
	else
		echo "!! populate $f FAILED ($OUT/populate-$f.log) — likely OOM; try a smaller KEYS or PARALLEL=0"
	fi
}

echo "== populate $KEYS keys/factory into $DIR ($BACKEND), GOMEMLIMIT=$GOMEMLIMIT =="
if [ "$PARALLEL" = 1 ]; then
	for f in $FACTORIES; do populate "$f" & done
	wait
else
	for f in $FACTORIES; do populate "$f"; done
fi

# 2. benchmark, one factory at a time, page cache dropped between for clean reads.
for f in $FACTORIES; do
	drop_caches
	echo ">> BlockWrite $f"
	go test "$PKG" -run='^$' -bench=BenchmarkDiskBlockWrite -timeout=2h \
		"${common[@]}" -disk-factory="$f" -disk-block="$BLOCK" -disk-warmup-ops="$WARMUP" -benchtime="${WRITE_N}x" \
		2>&1 | tee "$OUT/blockwrite-$f.txt"
	drop_caches
	echo ">> GetRandom $f"
	go test "$PKG" -run='^$' -bench=BenchmarkDiskGetRandom -timeout=2h \
		"${common[@]}" $read_flag -disk-factory="$f" -disk-warmup-ops="$WARMUP" -benchtime="${READ_N}x" \
		2>&1 | tee "$OUT/getrandom-$f.txt"
	drop_caches
	echo ">> GetMiss $f"
	go test "$PKG" -run='^$' -bench=BenchmarkDiskGetMiss -timeout=2h \
		"${common[@]}" $read_flag -disk-factory="$f" -disk-warmup-ops="$WARMUP" -benchtime="${READ_N}x" \
		2>&1 | tee "$OUT/getmiss-$f.txt"
done

echo
echo "==== summary (whole-run averages + tail-* steady state) ===="
grep -hE 'BenchmarkDisk(BlockWrite|GetRandom|GetMiss)' \
	"$OUT"/blockwrite-*.txt "$OUT"/getrandom-*.txt "$OUT"/getmiss-*.txt 2>/dev/null || true
echo
echo "==== convergence verdicts (steady iff |conv-ns-%| <= 5 read / 10 write AND |conv-reads-%| <= 2) ===="
grep -hE 'BenchmarkDisk(BlockWrite|GetRandom|GetMiss)' \
	"$OUT"/blockwrite-*.txt "$OUT"/getrandom-*.txt "$OUT"/getmiss-*.txt 2>/dev/null |
	awk '{
		name=$1; ns=""; rd=""; lim=5
		if (name ~ /BlockWrite/) lim=10
		for (i=2; i<NF; i++) {
			if ($(i+1)=="conv-ns-%") ns=$i
			if ($(i+1)=="conv-reads-%") rd=$i
		}
		if (ns=="" ) { print name": NO CONVERGENCE DATA (run too short for 2+ windows)"; next }
		bad=0
		if (ns<0?-ns>lim:ns>lim) bad=1
		if (rd!="" && (rd<0?-rd>2:rd>2)) bad=1
		print name": " (bad?"NOT CONVERGED":"CONVERGED") "  (conv-ns "ns"%, conv-reads "rd"%)"
	}'
echo "also eyeball the per-bench '--- BENCH' window series: flat tail = steady, monotone slide = still ramping"
echo "full output + populate logs in: $OUT/"
