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
# macOS ships shasum, minimal Linux images only sha256sum; same output format.
if command -v shasum >/dev/null 2>&1; then sha='shasum -a 256'; else sha=sha256sum; fi
: > "$out.tmp"
find testdata -name '*.go' | LC_ALL=C sort | while read -r f; do
	rel=${f#testdata/}
	case $rel in run/*|compile/*|errorcheck/*|gno/*) continue ;; esac
	case $rel in gotypes/*) up="$src/src/internal/types/testdata/${rel#gotypes/}" ;; *) up="$src/test/$rel" ;; esac
	[ -f "$up" ] || { echo "no upstream file for $rel" >&2; exit 1; }
	# CRLF becomes LF (one upstream file is a CRLF test); $(cat) drops every
	# trailing newline and printf puts one back: the normalisation the test
	# applies to a stripped body.
	sum=$(printf '%s\n' "$(tr -d '\r' < "$up")" | $sha | cut -d' ' -f1)
	printf '%s  %s\n' "$sum" "$rel" >> "$out.tmp"
done
mv "$out.tmp" "$out"
echo "$(wc -l < "$out" | tr -d ' ') entries -> $out"
