#!/bin/sh
# Writes UPSTREAM.sha256: one line per corpus file, the sha256 of its upstream
# body (trailing newlines normalised to one). Usage, from this directory:
#
#   ./upstream-manifest.sh /path/to/go-src-tree   # the release named in UPSTREAM
#
# testdata/<p> comes from $GOROOT/test/<p>; testdata/gotypes/<p> from
# $GOROOT/src/internal/types/testdata/<p>. Canary dirs (run, compile,
# errorcheck, gno) are not upstream and are skipped.
set -eu
src=${1:?go source tree}
cd "$(dirname "$0")"
out=UPSTREAM.sha256
: > "$out.tmp"
find testdata -name '*.go' | sort | while read -r f; do
	rel=${f#testdata/}
	case $rel in run/*|compile/*|errorcheck/*|gno/*) continue ;; esac
	case $rel in gotypes/*) up="$src/src/internal/types/testdata/${rel#gotypes/}" ;; *) up="$src/test/$rel" ;; esac
	[ -f "$up" ] || { echo "no upstream file for $rel" >&2; exit 1; }
	# $(cat) drops every trailing newline, printf puts one back: the same
	# normalisation the test applies to a stripped body.
	sum=$(printf '%s\n' "$(cat "$up")" | shasum -a 256 | cut -d' ' -f1)
	printf '%s  %s\n' "$sum" "$rel" >> "$out.tmp"
done
mv "$out.tmp" "$out"
echo "$(wc -l < "$out" | tr -d ' ') entries -> $out"
