#!/usr/bin/env bash
# Regression tests for the benchstat CSV cases this gate consumes.
set -euo pipefail

check=$(dirname "$0")/check.sh
tmp=$(mktemp -d)
trap 'rm -rf -- "$tmp"' EXIT

cat > "$tmp/clean.csv" <<'CSV'
,sec/op,CI,sec/op,CI,vs base,P
Stable-4,2e-08,0%,2e-08,0%,~,p=1.000 n=10
Added-4,,,2e-08,0%
geomean,2e-08,,2e-08,,+0.00%,
,allocs/op,CI,allocs/op,CI,vs base,P
Stable-4,0,0%,0,0%,~,p=1.000 n=10
CSV
"$check" "$tmp/clean.csv" > "$tmp/out"
grep -F 'UNCOMPARED Added-4 sec/op' "$tmp/out"

cat > "$tmp/regression.csv" <<'CSV'
,sec/op,CI,sec/op,CI,vs base,P
Slow-4,2e-08,0%,3e-08,0%,+50.00%,p=0.000 n=10
Missing-4,2e-08,0%
,allocs/op,CI,allocs/op,CI,vs base,P
Zero-4,0,0%,1,0%,?,p=0.000 n=10
CSV
if "$check" "$tmp/regression.csv" > "$tmp/out"; then
    echo 'benchci test: expected failure for regressions' >&2
    exit 1
fi
grep -F 'REGRESSION Slow-4 sec/op' "$tmp/out"
grep -F 'MISSING Missing-4 sec/op' "$tmp/out"
grep -F 'REGRESSION Zero-4 allocs/op' "$tmp/out"

: > "$tmp/empty.csv"
if "$check" "$tmp/empty.csv" > "$tmp/out"; then
    echo 'benchci test: expected failure for an empty comparison' >&2
    exit 1
fi
grep -F 'no comparable results' "$tmp/out"
