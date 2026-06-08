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
#   WRITE_N     BlockWrite benchtime, in blocks               [300]
#   READ_N      GetRandom benchtime, in ops                   [50000]
#   NODE_CACHE  in-process node LRU                           [10000]
#   FACTORIES   trees to run                                  ["iavl bptree"]
#   BUILD_BATCH keys per SaveVersion while building           [100000]
#               (lower to 25000 if the bptree populate OOMs)
#   PARALLEL    1 = populate both factories at once           [1]
#               (faster, but doubles RAM pressure; set 0 on a tight box)
#   DROP_CACHES 1 = drop OS page cache before each bench       [1] (needs sudo)
#   GOMEMLIMIT  Go soft heap cap, to survive big populates     [12GiB]
#   OUT         results/logs dir                               [./bench-out]

set -uo pipefail
cd "$(git rev-parse --show-toplevel)"

DIR="${DIR:-/data/bp32bench}"
KEYS="${KEYS:-100000000}"
BACKEND="${BACKEND:-pebbledb}"
BLOCK="${BLOCK:-1000}"
WRITE_N="${WRITE_N:-300}"
READ_N="${READ_N:-50000}"
NODE_CACHE="${NODE_CACHE:-10000}"
FACTORIES="${FACTORIES:-iavl bptree}"
BUILD_BATCH="${BUILD_BATCH:-100000}"
PARALLEL="${PARALLEL:-1}"
DROP_CACHES="${DROP_CACHES:-1}"
OUT="${OUT:-./bench-out}"
export GOMEMLIMIT="${GOMEMLIMIT:-12GiB}"

PKG=./tm2/pkg/bptree/benchmarks/
mkdir -p "$DIR" "$OUT"
common=(-disk-dir="$DIR" -disk-keys="$KEYS" -disk-backend="$BACKEND" -disk-node-cache="$NODE_CACHE")

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
		echo "!! populate $f FAILED ($OUT/populate-$f.log) — likely OOM; try BUILD_BATCH=25000, a smaller KEYS, or PARALLEL=0"
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
		"${common[@]}" -disk-factory="$f" -disk-block="$BLOCK" -benchtime="${WRITE_N}x" \
		2>&1 | tee "$OUT/blockwrite-$f.txt"
	drop_caches
	echo ">> GetRandom $f"
	go test "$PKG" -run='^$' -bench=BenchmarkDiskGetRandom -timeout=1h \
		"${common[@]}" -disk-factory="$f" -benchtime="${READ_N}x" \
		2>&1 | tee "$OUT/getrandom-$f.txt"
done

echo
echo "==== summary (reads/writes per op, ns) ===="
grep -hE 'BenchmarkDisk(BlockWrite|GetRandom)' \
	"$OUT"/blockwrite-*.txt "$OUT"/getrandom-*.txt 2>/dev/null || true
echo "full output + populate logs in: $OUT/"
