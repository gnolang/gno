#!/usr/bin/env bash
# Compare the benchmarks listed in suite.txt between two checkouts.
#
# Usage: run.sh <base checkout> <head checkout> <output dir>
#
# Each side's test binaries are compiled once, then run in COUNT interleaved
# rounds, alternating which side goes first, so that drift on a shared runner
# (thermal, noisy neighbours) lands on both sides instead of one. benchstat
# then compares the two samples; check.sh decides whether to fail.
set -euo pipefail

# base and head are read through ${!side}.
# shellcheck disable=SC2034
base=$(cd "$1" && pwd)
head=$(cd "$2" && pwd)
out=$(mkdir -p "$3" && cd "$3" && pwd)
count=${COUNT:-10}
suite=${SUITE:-$head/misc/benchci/suite.txt}

: >"$out/base.txt"
: >"$out/head.txt"
mkdir -p "$out/bin/base" "$out/bin/head"

# Read the suite: drop comments and blank lines.
mapfile -t lines < <(grep -v '^[[:space:]]*\(#\|$\)' "$suite")

# Compile every package once per side. A package missing on the base side
# (added by the pull request) is benchmarked on head only, and benchstat
# reports it without a comparison.
for i in "${!lines[@]}"; do
	read -r pkg _ _ <<<"${lines[$i]}"
	for side in base head; do
		dir=${!side}
		bin=$out/bin/$side/$i.test
		if [[ -d $dir/$pkg ]]; then
			(cd "$dir/$pkg" && go test -c -o "$bin" .) || {
				if [[ $side == head ]]; then exit 1; fi
				echo "benchci: $pkg does not build on base, head only" >&2
			}
		fi
	done
done

run() {
	local side=$1 i=$2 dir=${!1}
	local bin=$out/bin/$side/$i.test
	read -r pkg regexp benchtime <<<"${lines[$i]}"
	[[ -x $bin ]] || return 0
	# Run from the package directory: some benchmarks read testdata/.
	(cd "$dir/$pkg" && "$bin" -test.run '^$' -test.bench "$regexp" \
		-test.benchtime "$benchtime" -test.count 1 -test.benchmem) >>"$out/$side.txt"
}

start=$SECONDS
for ((r = 1; r <= count; r++)); do
	if ((r % 2)); then order=(base head); else order=(head base); fi
	for i in "${!lines[@]}"; do
		for side in "${order[@]}"; do run "$side" "$i"; done
	done
	echo "benchci: round $r/$count done at $((SECONDS - start))s" >&2
done

(cd "$out" && benchstat base.txt head.txt >benchstat.txt && benchstat -format csv base.txt head.txt >benchstat.csv)
cat "$out/benchstat.txt"
