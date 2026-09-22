#!/usr/bin/env bash
# render-upgrades.sh — render the table in UPGRADES.md from upgrades.json.
#
# upgrades.json is the source of truth for every binary the network has run;
# the table between the BEGIN/END GENERATED markers in UPGRADES.md is derived
# from it and never edited by hand. A version runs the blocks from its
# halt_height + 1 up to the next entry's halt_height; halt_height 0 is genesis.
#
# Usage:
#   ./render-upgrades.sh          # rewrite the table in place
#   ./render-upgrades.sh --check  # exit 1 if the table is stale (CI)
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

readonly LEDGER=upgrades.json
readonly DOC=UPGRADES.md
readonly BEGIN='<!-- BEGIN GENERATED (render-upgrades.sh) -->'
readonly END='<!-- END GENERATED -->'
readonly REPO=https://github.com/gnolang/gno

command -v jq >/dev/null 2>&1 || { echo "error: jq is required" >&2; exit 1; }
[[ -f ${LEDGER} && -f ${DOC} ]] || { echo "error: ${LEDGER} and ${DOC} must exist next to this script" >&2; exit 1; }
if ! grep -qxF "${BEGIN}" "${DOC}" || ! grep -qxF "${END}" "${DOC}"; then
	echo "error: ${DOC} has no '${BEGIN}' / '${END}' markers" >&2
	exit 1
fi

# One row per entry, in ledger order. Digests are shortened for the table; the
# full value stays in the JSON. Genesis has no halt and no proposal.
render_table() {
	printf '%s\n' "${BEGIN}"
	printf '| Version | Commit | Halt height | Halt time (UTC) | halt_min_version | GovDAO proposal | gnoland image digest | Consensus-relevant changes |\n'
	printf '|---|---|---|---|---|---|---|---|\n'
	jq -r --arg repo "${REPO}" '
		.upgrades[] |
		"| [\(.version)](\($repo)/releases/tag/\(.version))" +
		" | [\(.commit[0:9])](\($repo)/commit/\(.commit))" +
		" | \(if .halt_height == 0 then "genesis" else (.halt_height | tostring) end)" +
		" | \(.halt_time)" +
		" | \(if .halt_height == 0 then "—" elif .halt_min_version == "" then "*(empty)*" else "`\(.halt_min_version)`" end)" +
		" | \(if .proposal == null then "—" else "[#\(.proposal)](https://gno.land/r/gov/dao:\(.proposal))" end)" +
		" | \(if .image.digest == null then "*(pending)*" else "`\(.image.digest[0:19])…`" end)" +
		" | \(.changes | join(", ")) |"
	' "${LEDGER}"
	printf '%s\n' "${END}"
}

# The document with the generated block replaced, everything else untouched.
# The table goes through a file: BSD awk rejects a -v value containing newlines.
splice() {
	local table_file
	table_file="$(mktemp)"
	render_table >"${table_file}"
	awk -v table_file="${table_file}" -v begin="${BEGIN}" -v end="${END}" '
		$0 == begin {
			while ((getline line < table_file) > 0) print line
			close(table_file)
			skipping = 1
			next
		}
		$0 == end { skipping = 0; next }
		!skipping { print }
	' "${DOC}"
	rm -f "${table_file}"
}

case "${1-}" in
--check)
	if diff -u "${DOC}" <(splice); then
		echo "${DOC} is up to date with ${LEDGER}"
	else
		echo "error: ${DOC} is stale; run $(basename "${BASH_SOURCE[0]}") and commit the result" >&2
		exit 1
	fi
	;;
"")
	splice >"${DOC}.tmp"
	mv "${DOC}.tmp" "${DOC}"
	echo "${DOC} rendered from ${LEDGER}"
	;;
*)
	echo "usage: $(basename "${BASH_SOURCE[0]}") [--check]" >&2
	exit 2
	;;
esac
