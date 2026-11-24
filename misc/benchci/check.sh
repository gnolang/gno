#!/usr/bin/env bash
# Fail on significant regressions or missing head results for the gated metrics.
# Usage: check.sh <benchstat.csv>
set -euo pipefail

csv=$1
max_time=${MAX_TIME_REGRESSION:-20}
max_alloc=${MAX_ALLOC_REGRESSION:-10}

awk -F, -v max_time="$max_time" -v max_alloc="$max_alloc" '
	# A header names its metric and its time or allocation threshold.
	$1 == "" && $3 == "CI" {
		unit = $2
		limit = unit == "sec/op" ? max_time : unit == "allocs/op" ? max_alloc : -1
		next
	}
	$1 == "" || $1 == "geomean" || unit == "" || limit < 0 { next }
	{
		if ($2 == "") {
			printf "UNCOMPARED %s %s: no base result\n", $1, unit
			next
		}
		if ($4 == "") {
			printf "MISSING %s %s: no head result\n", $1, unit
			bad++
			next
		}
		compared++
		# benchstat prints "?" for a significant comparison whose base is zero.
		if ($6 == "?" && $2 + 0 == 0 && $4 + 0 > 0) {
			printf "REGRESSION %s %s: %s -> %s (%s)\n", $1, unit, $2, $4, $7
			bad++
			next
		}
		if ($6 !~ /^\+/ || $6 + 0 < limit) next
		printf "REGRESSION %s %s: %s (limit %s%%, %s)\n", $1, unit, $6, limit, $7
		bad++
	}
	END {
		if (!compared) { print "benchci: no comparable results"; exit 1 }
		if (bad) { printf "benchci: %d regression(s) or missing results\n", bad; exit 1 }
		print "benchci: no regression above the limits"
	}
' "$csv"
