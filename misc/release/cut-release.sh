#!/usr/bin/env bash
# cut-release.sh — tag a gno.land chain release, and check it is one.
#
# A release is a git tag, and almost everything that can go wrong with one is
# invisible at tagging time and expensive afterwards:
#
#   * The tag name does not parse as a version, so a halt proposal naming it
#     gates on byte equality instead of ordering — which refuses the very binary
#     the upgrade was cut for, and the chain cannot restart.
#   * The binaries do not carry the tag. `chain/mainnet`'s published gnoland
#     reports `develop`, because it was built without the ldflags; a node built
#     that way satisfies no halt_min_version at all.
#   * The tag lands on a commit that is not on master, so the chain is running
#     code that the development tree has never seen.
#
# Each of those is a preflight check below. The script does not push anything
# without --push, and prints every command it would run.
#
# Usage:
#   misc/release/cut-release.sh <version> [options]
#
# Options:
#   --chain <name>       chain branch to cut from       (default: mainnet)
#   --commit <ref>       commit to tag                  (default: the branch tip)
#   --previous <version> previous release, for the range (default: newest v* tag)
#   --halt-height <H>    also print the GovDAO halt proposal for a coordinated upgrade
#   --push               push the tag to origin (otherwise: dry run)
#   --allow-dirty        skip the clean-worktree check (local rehearsal only)
#
# Examples:
#   # what the mainnet launch should have been tagged with
#   misc/release/cut-release.sh v1.2.0 --commit 9c8eb132e
#
#   # a consensus-breaking upgrade, with the halt proposal to go with it
#   misc/release/cut-release.sh v1.3.0 --halt-height 120000 --push

set -euo pipefail

readonly REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# Kept identical to .github/workflows/release-chain-tag.yml: the point of the
# build check below is to prove the workflow's flags produce a binary carrying
# the tag, which it cannot do if it builds with different ones.
# gno.land/pkg/gnoland.TestReleaseToolingMatchesTheParser reads both out of the
# files and fails if they drift, so keep each on one line.
readonly VERSION_PKG="github.com/gnolang/gno/tm2/pkg/version.Version"

# The shape gno.land/pkg/gnoland.parseReleaseVersion accepts, spelled as an ERE:
# semver's own grammar, so that leading zeros, signed components and a dangling
# "-" are refused here rather than at the halt. The same test holds this to the
# Go parser over a corpus, so keep it on one line too.
readonly VERSION_RE='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*)?$'

CHAIN="mainnet"
COMMIT=""
PREVIOUS=""
HALT_HEIGHT=""
PUSH=0
ALLOW_DIRTY=0
VERSION=""
# A checkout of the commit being tagged, shared by every preflight check that
# has to inspect that tree rather than the one the operator is standing on. A
# single EXIT trap rather than a RETURN one: a RETURN trap set inside a function
# fires again when main returns, where the local it referred to is gone and
# `set -u` turns cleanup into an error.
SCRATCH=""
WORKTREE=""
cleanup() {
	if [[ -n ${WORKTREE} ]]; then
		git -C "${REPO_ROOT}" worktree remove --force "${WORKTREE}" >/dev/null 2>&1 || true
	fi
	[[ -z ${SCRATCH} ]] || rm -rf "${SCRATCH}"
}
trap cleanup EXIT

die() {
	printf '\033[31merror:\033[0m %s\n' "$*" >&2
	exit 1
}
info() { printf '\033[36m==>\033[0m %s\n' "$*"; }
ok() { printf '  \033[32m✓\033[0m %s\n' "$*"; }
warn() { printf '  \033[33m!\033[0m %s\n' "$*" >&2; }
# Up to the first blank line, so adding an option to the header above cannot
# silently truncate --help.
usage() { sed -n '2,/^$/p' "${BASH_SOURCE[0]}" | sed 's|^# \{0,1\}||'; }

parse_args() {
	while [[ $# -gt 0 ]]; do
		case "$1" in
		-h | --help)
			usage
			exit 0
			;;
		--chain)
			[[ -n ${2-} ]] || die "--chain needs a value (e.g. mainnet)"
			CHAIN="$2"
			shift 2
			;;
		--commit)
			[[ -n ${2-} ]] || die "--commit needs a ref"
			COMMIT="$2"
			shift 2
			;;
		--previous)
			[[ -n ${2-} ]] || die "--previous needs a version (e.g. v1.2.0)"
			PREVIOUS="$2"
			shift 2
			;;
		--halt-height)
			# Not just "non-empty": 0 is the documented *cancel* form (see
			# NewSetHaltRequest in examples/gno.land/r/sys/params/halt.gno), so
			# accepting it here would emit a proposal that cancels the halt while
			# every other artifact describes scheduling one. A bare --halt-height
			# followed by another flag would otherwise swallow it, too.
			[[ ${2-} =~ ^[1-9][0-9]*$ ]] \
				|| die "--halt-height takes a positive block height (got ${2-<missing>}).
       Height 0 cancels a scheduled halt; emit that one by hand."
			HALT_HEIGHT="$2"
			shift 2
			;;
		--push)
			PUSH=1
			shift
			;;
		--allow-dirty)
			ALLOW_DIRTY=1
			shift
			;;
		-*) die "unknown option $1 (try --help)" ;;
		*)
			[[ -z ${VERSION} ]] || die "unexpected argument $1"
			VERSION="$1"
			shift
			;;
		esac
	done
	[[ -n ${VERSION} ]] || {
		usage
		exit 1
	}
}

# ---------------------------------------------------------------- preflight

# A tag outside VERSION_RE cannot be used as halt_min_version, which is most of
# the point of tagging.
check_version_shape() {
	[[ ${VERSION} =~ ${VERSION_RE} ]] \
		|| die "version must be vMAJOR.MINOR.PATCH, optionally -prerelease (got ${VERSION}).
       This is the shape the node parses for halt_min_version; 'chain/<name>'
       tags do not order and cannot gate an upgrade. See RELEASING.md."
	ok "version shape: ${VERSION}"
}

check_worktree() {
	[[ -d ${REPO_ROOT}/.git || -f ${REPO_ROOT}/.git ]] || die "not a git checkout: ${REPO_ROOT}"
	if [[ ${ALLOW_DIRTY} -eq 0 ]] && [[ -n "$(git -C "${REPO_ROOT}" status --porcelain)" ]]; then
		die "worktree is dirty; commit or stash first (or pass --allow-dirty to rehearse)"
	fi
	ok "worktree clean"
}

check_tag_free() {
	if git -C "${REPO_ROOT}" rev-parse -q --verify "refs/tags/${VERSION}" >/dev/null; then
		die "tag ${VERSION} already exists locally. Tags are immutable — pick the next version."
	fi
	# ls-remote exits 2 when no ref matches, and anything else (128 for an
	# unreachable remote) means the question was never answered. Only 2 means
	# free: reading every failure as free would skip the check without saying so.
	local rc=0
	git -C "${REPO_ROOT}" ls-remote --exit-code --tags origin "refs/tags/${VERSION}" >/dev/null 2>&1 || rc=$?
	case ${rc} in
	0) die "tag ${VERSION} already exists on origin. Tags are immutable — pick the next version." ;;
	2) ;;
	*) die "could not ask origin whether ${VERSION} exists (git ls-remote exited ${rc}).
       Check the remote and network before cutting." ;;
	esac
	ok "tag ${VERSION} is free"
}

resolve_commit() {
	local branch="chain/${CHAIN}"
	if [[ -z ${COMMIT} ]]; then
		COMMIT="$(git -C "${REPO_ROOT}" rev-parse "origin/${branch}" 2>/dev/null)" \
			|| die "no origin/${branch}; pass --chain or --commit"
	fi
	COMMIT="$(git -C "${REPO_ROOT}" rev-parse "${COMMIT}^{commit}")" || die "cannot resolve --commit"

	git -C "${REPO_ROOT}" merge-base --is-ancestor "${COMMIT}" "origin/${branch}" 2>/dev/null \
		|| warn "${COMMIT:0:9} is not on origin/${branch} — is --chain right?"

	ok "tagging ${COMMIT:0:9} ($(git -C "${REPO_ROOT}" log -1 --format='%s' "${COMMIT}" | cut -c1-60))"
}

# A chain branch may lag master, never lead it: master is the development tree
# and every line the chain runs has to exist there. A tag on a commit master has
# never seen means the chain is running code nobody can review on master.
check_on_master() {
	if git -C "${REPO_ROOT}" merge-base --is-ancestor "${COMMIT}" origin/master 2>/dev/null; then
		ok "commit is contained in origin/master"
		return
	fi
	# Trees first: chain branches take merges from master and give code back
	# through squashed PRs, so commit identity is noisy where content is exact.
	if git -C "${REPO_ROOT}" diff --quiet origin/master "${COMMIT}" 2>/dev/null; then
		ok "tree is identical to origin/master"
		return
	fi

	local missing
	missing="$(git -C "${REPO_ROOT}" log --oneline --no-merges -n 20 origin/master.."${COMMIT}" -- \
		":!misc/deployments" 2>/dev/null)" || {
		warn "cannot compare against origin/master (is it fetched?); skipping the check"
		return
	}
	if [[ -z ${missing} ]]; then
		ok "differs from origin/master only under misc/deployments"
		return
	fi
	warn "these commits are not on origin/master and are not deployment-only:"
	printf '%s\n' "${missing}" | sed 's/^/      /' >&2
	warn "check each one: a commit squash-merged to master shows up here even"
	warn "though its content landed. Anything genuinely missing should be ported"
	warn "to master first, or the chain runs code the development tree never saw."
}

# The ledger is what operators read to know which version to run and what a
# replaying node follows; a release without an entry is invisible to both. A
# warning, not a refusal: the entry may legitimately land after the tag. The
# pre-release suffix is stripped because an rc rehearses the final version's
# entry rather than getting one of its own.
check_ledger_entry() {
	local ledger="${REPO_ROOT}/misc/deployments/${CHAIN}.gno.land/upgrades.json"
	[[ -f ${ledger} ]] || return 0
	if ! command -v jq >/dev/null 2>&1; then
		warn "jq not found; skipping the upgrades.json check"
		return 0
	fi
	local final="${VERSION%%-*}"
	if jq -e --arg v "${final}" '.upgrades[] | select(.version == $v)' "${ledger}" >/dev/null 2>&1; then
		ok "upgrades.json has an entry for ${final}"
	else
		warn "${ledger#"${REPO_ROOT}"/} has no entry for ${final}; add it and render UPGRADES.md"
		warn "   (misc/deployments/${CHAIN}.gno.land/render-upgrades.sh) before announcing the release"
	fi
}

# Everything below inspects the commit being tagged, which is usually not the
# commit the operator has checked out: the default is origin/chain/<name>, and
# drift that exists only there is drift the validators would be running.
prepare_worktree() {
	SCRATCH="$(mktemp -d)"
	WORKTREE="${SCRATCH}/src"
	git -C "${REPO_ROOT}" worktree add --detach "${WORKTREE}" "${COMMIT}" >/dev/null 2>&1 \
		|| die "could not create a worktree at ${COMMIT:0:9}"
}

check_protocol_constants() {
	# This checkout's script against the tagged commit's tree: the commit may
	# predate misc/release/ entirely, which the v1.2.0 example above does.
	local out
	out="$(GNO_REPO_ROOT="${WORKTREE}" \
		"${REPO_ROOT}/misc/release/bump-protocol-version.sh" --check 2>&1)" || die \
		"the protocol-version constants at ${COMMIT:0:9} did not verify:

$(printf '%s\n' "${out}" | sed 's/^/       /')

       If they genuinely disagree, misc/release/bump-protocol-version.sh moves
       all six together."
	# Warnings on the success path would otherwise vanish into the captured
	# output — notably "the invariant test is not in this tree", which is how a
	# check that verified less than it claims stays quiet.
	# bump-protocol-version.sh colours its output, so strip ANSI first:
	# grep '^warning:' never fires against "\033[33mwarning:\033[0m".
	local warnings
	warnings="$(printf '%s\n' "${out}" | sed -e $'s/\033\\[[0-9;]*m//g' | grep '^warning:' || true)"
	if [[ -n ${warnings} ]]; then
		while IFS= read -r line; do warn "${line#warning: }"; done <<<"${warnings}"
	fi
	ok "protocol-version constants agree at ${COMMIT:0:9}"
}

# The check that would have caught chain/mainnet's binaries reporting `develop`:
# build the way the release workflow builds, then ask the binary what it is.
check_build_reports_tag() {
	info "building gnoland at ${COMMIT:0:9} to verify the version is compiled in"

	local build_log="${SCRATCH}/build.log"
	(cd "${WORKTREE}" && go build \
		-ldflags "-w -s -X ${VERSION_PKG}=${VERSION}" \
		-o "${SCRATCH}/gnoland" ./gno.land/cmd/gnoland) >"${build_log}" 2>&1 \
		|| die "gnoland does not build at ${COMMIT:0:9}:

$(tail -20 "${build_log}" | sed 's/^/       /')"

	# Captured rather than piped: under `set -o pipefail` a binary that exits
	# non-zero — which is exactly what a partially applied protocol-version bump
	# does, in an init() guard — would abort the script here, so the message
	# below would never reach the operator and neither would the panic.
	local out reported
	out="$("${SCRATCH}/gnoland" version 2>&1)" \
		|| die "the binary built at ${COMMIT:0:9} failed to run:

$(printf '%s\n' "${out}" | sed 's/^/       /')"

	# Match the version line itself: mdbx writes debug lines around it.
	reported="$(printf '%s\n' "${out}" | awk '$1 == "gnoland" && $2 == "version:" {print $3}')"
	[[ ${reported} == "${VERSION}" ]] \
		|| die "the built binary reports ${reported:-<nothing>}, not ${VERSION}.
       Release binaries must carry the tag or they satisfy no halt_min_version.
       Check the -ldflags in .github/workflows/release-chain-tag.yml."
	ok "binary reports ${reported}"
}

# ------------------------------------------------------------ classification

newest_release_tag() {
	# Reachable from the commit being tagged, so a tag on an unrelated line is
	# not mistaken for this release's predecessor, and final releases only: a
	# pre-release is not the release it leads to, and picking v1.3.0-rc.1 as the
	# predecessor of v1.3.0 makes a coordinated upgrade classify as a PATCH.
	local tags
	tags="$(git -C "${REPO_ROOT}" tag --list 'v*' --sort=-v:refname --merged "${COMMIT}" 2>/dev/null)" \
		|| return 0
	# awk rather than head: head exits early, which SIGPIPEs git and trips pipefail.
	printf '%s\n' "${tags}" | awk 'NF && !/-/ {print; exit}'
}

# Says what kind of upgrade this is, because the answer decides whether
# validators have to be coordinated — and that is not a git question.
classify() {
	[[ -n ${PREVIOUS} ]] || PREVIOUS="$(newest_release_tag)"
	if [[ -z ${PREVIOUS} ]]; then
		info "no previous v* tag; treating ${VERSION} as the first of its line"
		return
	fi
	if ! git -C "${REPO_ROOT}" rev-parse -q --verify "${PREVIOUS}^{commit}" >/dev/null; then
		warn "--previous ${PREVIOUS} does not resolve; skipping the change summary"
		PREVIOUS=""
		return
	fi

	# Strip any pre-release suffix: v1.3.0-rc.1 and v1.3.0 are the same line, and
	# --previous may name one even though newest_release_tag will not.
	local prev_mm="${PREVIOUS#v}" new_mm="${VERSION#v}"
	prev_mm="${prev_mm%%-*}"
	new_mm="${new_mm%%-*}"
	local prev_major="${prev_mm%%.*}" new_major="${new_mm%%.*}"
	local prev_minor new_minor
	prev_minor="$(printf '%s' "${prev_mm}" | cut -d. -f2)"
	new_minor="$(printf '%s' "${new_mm}" | cut -d. -f2)"

	info "since ${PREVIOUS}:"
	git -C "${REPO_ROOT}" log --oneline -n 30 "${PREVIOUS}..${COMMIT}" | sed 's/^/      /'
	local n
	n="$(git -C "${REPO_ROOT}" rev-list --count "${PREVIOUS}..${COMMIT}")"
	[[ ${n} -le 30 ]] || printf '      ... and %d more\n' "$((n - 30))"

	# Anything the contributors marked breaking is a coordinated upgrade
	# regardless of how the version was numbered, so say so rather than trusting
	# the number the operator typed.
	local breaking
	breaking="$(git -C "${REPO_ROOT}" log --format='%s%n%b' "${PREVIOUS}..${COMMIT}" \
		| grep -cE '^[a-z]+(\([^)]*\))?!:|^BREAKING' || true)"

	printf '\n'
	if [[ ${new_major} != "${prev_major}" ]]; then
		info "MAJOR bump: a new network, incompatible genesis. See RELEASING.md."
	elif [[ ${new_minor} != "${prev_minor}" ]]; then
		info "MINOR bump: coordinated upgrade. Validators must halt together;"
		info "    pass --halt-height to emit the GovDAO proposal."
	else
		info "PATCH: no validator coordination needed."
		if [[ ${breaking} -gt 0 ]]; then
			warn "...but ${breaking} commit(s) in this range are marked breaking (feat!:/BREAKING)."
			warn "   A consensus change is a MINOR bump, not a patch. Re-check the number."
		fi
	fi
	printf '\n'
}

# ------------------------------------------------------------------- output

# Printed, not written. transactions/migration/ is a replay archive whose
# entries are meta.json directories read by name (see gen-genesis.sh), so a bare
# .gno dropped in there has no reader — and a file written before the tag is a
# file left behind when the tag is undone. The proposal is a governance action,
# and misc/govdao-scripts/set-halt.sh is what performs it.
print_halt_proposal() {
	# The wrapper is named govdao-exec.sh in most deployments and govdao in a
	# couple of older ones; both dispatch into misc/govdao-scripts/.
	local wrapper=""
	local dir name
	for dir in "misc/deployments/${CHAIN}.gno.land" "misc/deployments/${CHAIN}"; do
		for name in govdao-exec.sh govdao; do
			if [[ -x "${REPO_ROOT}/${dir}/${name}" ]]; then
				wrapper="${dir}/${name}"
				break 2
			fi
		done
	done
	[[ -n ${wrapper} ]] || die "no govdao wrapper under misc/deployments for --chain ${CHAIN}"

	info "coordinated upgrade: the GovDAO proposal for it is"
	printf '\n    %s set-halt %s %s\n\n' "${wrapper}" "${HALT_HEIGHT}" "${VERSION}"
	info "which creates this proposal:"
	cat <<EOF

    // Every node stops after committing block ${HALT_HEIGHT}, and refuses to
    // restart on a binary older than ${VERSION}.
    package main

    import (
    	"gno.land/r/gov/dao"
    	"gno.land/r/sys/params"
    )

    func main(cur realm) {
    	req := params.NewSetHaltRequest(cross(cur), ${HALT_HEIGHT}, "${VERSION}")
    	dao.MustCreateProposal(cross(cur), req)
    }

EOF
	warn "this creates the proposal only — members vote and execute separately"
	warn "before it passes, every operator should confirm halt_height is unset in"
	warn "their own config.toml. See gno.land/cmd/gnoland/UPGRADES.md."
}

tag_message() {
	printf '%s %s\n\nReleased from chain/%s at %s.' \
		"${CHAIN}" "${VERSION}" "${CHAIN}" "${COMMIT:0:9}"
	[[ -z ${PREVIOUS} ]] || printf '\nPrevious release: %s.' "${PREVIOUS}"
	[[ -z ${HALT_HEIGHT} ]] || printf '\nCoordinated upgrade at height %s, halt_min_version %s.' \
		"${HALT_HEIGHT}" "${VERSION}"
	printf '\n'
}

create_tag() {
	git -C "${REPO_ROOT}" tag -a "${VERSION}" "${COMMIT}" -m "$(tag_message)"
	ok "created annotated tag ${VERSION}"
}

main() {
	parse_args "$@"

	info "preflight"
	check_version_shape
	check_worktree
	check_tag_free
	resolve_commit
	check_on_master
	check_ledger_entry
	prepare_worktree
	check_protocol_constants
	check_build_reports_tag
	printf '\n'

	classify

	[[ -z ${HALT_HEIGHT} ]] || print_halt_proposal

	# A dry run changes nothing, not even locally: a tag created here would have
	# to be remembered and deleted, and until it was, check_tag_free would refuse
	# the real run as if the version had already been released.
	if [[ ${PUSH} -eq 0 ]]; then
		printf '\n'
		info "dry run — nothing was created. The tag would be:"
		printf '\n%s\n\n' "$(tag_message | sed 's/^/    /')"
		info "to cut it for real, re-run with --push"
		return
	fi

	create_tag
	printf '\n'
	info "pushing"
	git -C "${REPO_ROOT}" push origin "refs/tags/${VERSION}:refs/tags/${VERSION}"
	ok "pushed ${VERSION} — release / chain-tag will build and attach the binaries"
}

main "$@"
