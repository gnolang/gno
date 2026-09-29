#!/usr/bin/env bash
# bump-protocol-version.sh — move the six protocol-version constants together.
#
# The protocol version is one value that lives in six files, split that way to
# avoid import cycles rather than because the parts can differ. Each file guards
# the next in an init(), so a bump that misses one panics every node at startup:
#
#   panic: protocol version mismatch: bft is v1.1.0 but abci is v1.0.0-rc.0
#
# That is a loud failure, but it is found at run time by whoever starts the node,
# which on a release day is a validator. This script edits all six, then proves
# they agree by running the test that pins the invariant.
#
# This is NOT the release version. `tm2/pkg/version.Version` is compiled in from
# the git tag and is what halt_min_version compares; see cut-release.sh. The
# constants here are the protocol version negotiated with peers, and
# VersionSet.CompatibleWith refuses a peer whose MAJOR differs — so bumping the
# major partitions the network and is a coordinated upgrade in its own right.
#
# Usage:
#   misc/release/bump-protocol-version.sh <version>
#   misc/release/bump-protocol-version.sh --check
#
# Example:
#   misc/release/bump-protocol-version.sh v1.0.0

set -euo pipefail

# GNO_REPO_ROOT lets cut-release.sh point this script at a worktree of the
# commit being tagged rather than the checkout it was invoked from. The commit
# may predate this script, so the script has to travel to the tree, not the
# other way round.
readonly REPO_ROOT="${GNO_REPO_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"

# file:constant pairs. Keep in sync with TestProtocolVersionsAgree in
# tm2/pkg/bft/version/version_test.go, which is what fails if this list rots.
readonly CONSTANTS=(
	"tm2/pkg/crypto/version.go:Version"
	"tm2/pkg/bft/abci/version/version.go:Version"
	"tm2/pkg/bft/blockchain/version/version.go:Version"
	"tm2/pkg/bft/types/version/version.go:BlockVersion"
	"tm2/pkg/p2p/version/version.go:Version"
	"tm2/pkg/bft/version/version.go:Version"
)

die() {
	printf '\033[31merror:\033[0m %s\n' "$*" >&2
	exit 1
}

info() { printf '\033[36m==>\033[0m %s\n' "$*"; }
warn() { printf '\033[33mwarning:\033[0m %s\n' "$*" >&2; }

# The declaration we rewrite, for either `const X = "..."`, `const X string =
# "..."` or a bare `X = "..."` inside a var block. The ^ anchor is load-bearing:
# it keeps the substitution off the init() guard lines, which mention the same
# identifiers qualified by a package name. '#' is the sed delimiter because the
# pattern itself contains '|'.
decl_re() {
	printf '^([[:space:]]*(const|var)?[[:space:]]*%s([[:space:]]+string)?[[:space:]]*=[[:space:]]*)"([^"]*)"' "$1"
}

# read_constant <file> <name> — print the current value. Fails unless exactly one
# declaration matches, so a second `Version = "..."` appearing in one of these
# files is caught here rather than silently half-patched.
read_constant() {
	local file="$1" name="$2" matches
	matches="$(sed -nE "s#$(decl_re "${name}").*#\4#p" "${REPO_ROOT}/${file}")"
	[[ -n ${matches} ]] || die "no ${name} declaration in ${file} — has the file moved?"
	[[ $(printf '%s\n' "${matches}" | wc -l) -eq 1 ]] \
		|| die "${file} has more than one ${name} declaration; refusing to guess"
	printf '%s' "${matches}"
}

# write_constant <file> <name> <value>
write_constant() {
	local file="$1" name="$2" value="$3"
	sed -i.bak -E "s#$(decl_re "${name}")#\1\"${value}\"#" "${REPO_ROOT}/${file}"
	rm -f "${REPO_ROOT}/${file}.bak"
}

# current_version — the value all six are expected to share, taken from the
# canonical one. Errors if they already disagree.
current_version() {
	local seen="" file name value
	for entry in "${CONSTANTS[@]}"; do
		file="${entry%%:*}"
		name="${entry##*:}"
		value="$(read_constant "${file}" "${name}")"
		[[ -n ${value} ]] || die "could not read ${name} from ${file} — has the file moved?"
		if [[ -z ${seen} ]]; then
			seen="${value}"
		elif [[ ${value} != "${seen}" ]]; then
			die "constants already disagree: ${file} has ${value}, expected ${seen}. Fix that first."
		fi
	done
	printf '%s' "${seen}"
}

verify() {
	info "verifying the constants agree"
	# `go test -run` exits 0 when the pattern matches nothing, so a renamed or
	# moved test would turn this whole check into a no-op without saying so.
	# -list answers whether it is there before -run is asked to run it.
	local listed
	listed="$(cd "${REPO_ROOT}" && go test ./tm2/pkg/bft/version/ \
		-list '^TestProtocolVersionsAgree$' 2>/dev/null | grep -cx 'TestProtocolVersionsAgree')" || true
	if [[ ${listed} -eq 0 ]]; then
		warn "TestProtocolVersionsAgree is not in ${REPO_ROOT}/tm2/pkg/bft/version —"
		warn "renamed, moved, or a tree predating it. The six constants were compared"
		warn "by reading the files; the init() guards were not exercised."
		return
	fi
	(cd "${REPO_ROOT}" && go test ./tm2/pkg/bft/version/ -run '^TestProtocolVersionsAgree$' -count=1) \
		|| die "TestProtocolVersionsAgree failed — the constants are out of sync"
}

main() {
	local target="${1-}"

	case "${target}" in
	"" | -h | --help)
		sed -n '2,26p' "${BASH_SOURCE[0]}" | sed 's|^# \{0,1\}||'
		exit 0
		;;
	--check)
		local now
		now="$(current_version)"
		info "protocol version is ${now}"
		verify
		exit 0
		;;
	esac

	# The same ERE cut-release.sh uses, for the same reason: versionset compares
	# these with semver.Compare, so a value semver will not parse orders wrong.
	# gno.land/pkg/gnoland.TestReleaseToolingMatchesTheParser holds both to the
	# Go parser, so keep this on one line.
	[[ ${target} =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*)?$ ]] \
		|| die "version must look like v1.2.3 or v1.2.3-rc.1, got ${target}"

	local now
	now="$(current_version)"

	if [[ ${now} == "${target}" ]]; then
		info "already at ${target}; nothing to do"
		exit 0
	fi

	# A major bump is a network partition, not a version bump: peers that
	# disagree on the major are refused by VersionSet.CompatibleWith, so old and
	# new nodes cannot gossip at all. Say so before it happens.
	local now_major="${now%%.*}" target_major="${target%%.*}"
	if [[ ${now_major} != "${target_major}" ]]; then
		warn "MAJOR bump ${now_major} -> ${target_major}: nodes on the old major will not"
		warn "peer with nodes on the new one. This must ride a coordinated halt, and"
		warn "every validator must switch at the same height. See RELEASING.md."
	fi

	info "bumping protocol version ${now} -> ${target}"
	# A failure part-way leaves the tree with the constants disagreeing, which is
	# the state current_version() refuses to start from. Name what was already
	# written so the next run is a `git checkout` away rather than a hunt.
	local file name patched=()
	for entry in "${CONSTANTS[@]}"; do
		file="${entry%%:*}"
		name="${entry##*:}"
		write_constant "${file}" "${name}" "${target}" || die "failed to patch ${name} in ${file}.
       Already patched to ${target}: ${patched[*]:-<none>}
       Revert with: git -C ${REPO_ROOT} checkout -- ${patched[*]:-}"
		local got
		got="$(read_constant "${file}" "${name}")"
		[[ ${got} == "${target}" ]] || die "failed to patch ${name} in ${file} (still ${got}).
       Already patched to ${target}: ${patched[*]:-<none>}
       Revert with: git -C ${REPO_ROOT} checkout -- ${patched[*]:-}"
		patched+=("${file}")
		printf '    %s\n' "${file}"
	done

	local unformatted
	unformatted="$(cd "${REPO_ROOT}" && gofmt -l "${CONSTANTS[@]%%:*}")"
	[[ -z ${unformatted} ]] || die "gofmt would reformat the patched files: ${unformatted}"

	verify

	info "done. Review the diff, then commit:"
	printf '\n    git add %s\n' "$(printf '%s ' "${CONSTANTS[@]%%:*}")"
	printf "    git commit -m 'chore(tm2)!: bump the protocol version to %s'\n\n" "${target}"
}

main "$@"
