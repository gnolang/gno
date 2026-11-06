#!/usr/bin/env bash
# Fail when benchstat reports a large, statistically significant regression.
#
# Usage: check.sh <benchstat.csv>
#
# A row fails only if both hold:
#   - benchstat found the difference significant (p < 0.05, Mann-Whitney U;
#     otherwise it prints "~" instead of a delta), and
#   - the slowdown is at least MAX_TIME_REGRESSION percent for sec/op, or
#     MAX_ALLOC_REGRESSION percent for allocs/op.
# CI runners are noisy, so the time floor is wide on purpose: this gate is for
# the 2x slowdown nobody noticed, not for a 3% drift. allocs/op does not depend
# on the machine, so it gets a tighter floor.
set -euo pipefail

csv=$1
max_time=${MAX_TIME_REGRESSION:-20}
max_alloc=${MAX_ALLOC_REGRESSION:-10}

awk -F, -v max_time="$max_time" -v max_alloc="$max_alloc" '
	# A table header names its unit in the second column: ",sec/op,CI,...".
	$1 == "" && $3 == "CI" { unit = $2; next }
	$1 == "" || $1 == "geomean" || NF < 7 { next }
	{
		delta = $6
		if (delta == "~" || delta !~ /^\+/) next
		sub(/^\+/, "", delta); sub(/%$/, "", delta)
		limit = unit == "sec/op" ? max_time : unit == "allocs/op" ? max_alloc : -1
		if (limit < 0 || delta + 0 < limit) next
		printf "REGRESSION %s %s: +%s%% (limit %s%%, %s)\n", $1, unit, delta, limit, $7
		bad++
	}
	END {
		if (bad) { printf "benchci: %d regression(s) above the limits\n", bad; exit 1 }
		print "benchci: no regression above the limits"
	}
' "$csv"
