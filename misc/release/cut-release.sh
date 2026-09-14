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
#   --halt-height <H>    also emit a GovDAO halt proposal for a coordinated upgrade
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
readonly VERSION_PKG="github.com/gnolang/gno/tm2/pkg/version.Version"

CHAIN="mainnet"
COMMIT=""
PREVIOUS=""
HALT_HEIGHT=""
PUSH=0
ALLOW_DIRTY=0
VERSION=""
# Scratch dir for the build check. A single EXIT trap rather than a RETURN one:
# a RETURN trap set inside a function fires again when main returns, where the
# local it referred to is gone and `set -u` turns cleanup into an error.
SCRATCH=""
cleanup() { [[ -z ${SCRATCH} ]] || rm -rf "${SCRATCH}"; }
trap cleanup EXIT

die() {
	printf '\033[31merror:\033[0m %s\n' "$*" >&2
	exit 1
}
info() { printf '\033[36m==>\033[0m %s\n' "$*"; }
ok() { printf '  \033[32m✓\033[0m %s\n' "$*"; }
warn() { printf '  \033[33m!\033[0m %s\n' "$*" >&2; }
usage() { sed -n '2,34p' "${BASH_SOURCE[0]}" | sed 's|^# \{0,1\}||'; }

parse_args() {
	while [[ $# -gt 0 ]]; do
		case "$1" in
		-h | --help)
			usage
			exit 0
			;;
		--chain)
			CHAIN="${2-}"
			shift 2
			;;
		--commit)
			COMMIT="${2-}"
			shift 2
			;;
		--previous)
			PREVIOUS="${2-}"
			shift 2
			;;
		--halt-height)
			HALT_HEIGHT="${2-}"
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

# The shape gno.land/pkg/gnoland.parseReleaseVersion accepts. A tag outside it
# cannot be used as halt_min_version, which is most of the point of tagging.
check_version_shape() {
	[[ ${VERSION} =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]] \
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
	if git -C "${REPO_ROOT}" ls-remote --exit-code --tags origin "refs/tags/${VERSION}" >/dev/null 2>&1; then
		die "tag ${VERSION} already exists on origin. Tags are immutable — pick the next version."
	fi
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
		":!misc/deployments")"
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

check_protocol_constants() {
	"${REPO_ROOT}/misc/release/bump-protocol-version.sh" --check >/dev/null 2>&1 \
		|| die "the protocol-version constants disagree; run misc/release/bump-protocol-version.sh"
	ok "protocol-version constants agree"
}

# The check that would have caught chain/mainnet's binaries reporting `develop`:
# build the way the release workflow builds, then ask the binary what it is.
check_build_reports_tag() {
	SCRATCH="$(mktemp -d)"
	local tmp="${SCRATCH}"

	info "building gnoland at ${COMMIT:0:9} to verify the version is compiled in"
	local worktree="${tmp}/src"
	git -C "${REPO_ROOT}" worktree add --detach "${worktree}" "${COMMIT}" >/dev/null 2>&1 \
		|| die "could not create a worktree at ${COMMIT:0:9}"

	local built=1
	if (cd "${worktree}" && go build -trimpath \
		-ldflags "-w -s -X ${VERSION_PKG}=${VERSION}" \
		-o "${tmp}/gnoland" ./gno.land/cmd/gnoland >/dev/null 2>&1); then
		built=0
	fi

	git -C "${REPO_ROOT}" worktree remove --force "${worktree}" >/dev/null 2>&1 || true
	[[ ${built} -eq 0 ]] || die "gnoland does not build at ${COMMIT:0:9}"

	local reported
	# Match the version line itself: mdbx writes debug lines around it.
	reported="$(GNOROOT="${REPO_ROOT}" "${tmp}/gnoland" version 2>/dev/null \
		| awk '$1 == "gnoland" && $2 == "version:" {print $3}')"
	[[ ${reported} == "${VERSION}" ]] \
		|| die "the built binary reports ${reported:-<nothing>}, not ${VERSION}.
       Release binaries must carry the tag or they satisfy no halt_min_version.
       Check the -ldflags in .github/workflows/release-chain-tag.yml."
	ok "binary reports ${reported}"
}

# ------------------------------------------------------------ classification

newest_release_tag() {
	# awk over head: head exits early, which SIGPIPEs git and trips pipefail.
	git -C "${REPO_ROOT}" tag --list 'v*' --sort=-v:refname | awk 'NR==1'
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

	local prev_mm="${PREVIOUS#v}" new_mm="${VERSION#v}"
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

emit_halt_proposal() {
	local dir="${REPO_ROOT}/misc/deployments/${CHAIN}.gno.land/transactions/migration/halt-${VERSION}"
	[[ -d "${REPO_ROOT}/misc/deployments/${CHAIN}.gno.land" ]] \
		|| die "no misc/deployments/${CHAIN}.gno.land — pass the right --chain"

	mkdir -p "${dir}"
	cat >"${dir}/halt_${VERSION//[.-]/_}.gno" <<EOF
// Coordinated halt for the ${VERSION} upgrade.
//
// Every node stops after committing block ${HALT_HEIGHT}, and refuses to
// restart on a binary older than ${VERSION}. Generated by
// misc/release/cut-release.sh; the version string is the release tag, which
// gno.land/pkg/gnoland.meetsMinVersion parses and orders.
//
// Before passing this: every operator should confirm halt_height is unset in
// their own config.toml. A future-dated value there halts the node at the wrong
// block. See RELEASING.md and gno.land/cmd/gnoland/README.md.
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
	ok "wrote ${dir#"${REPO_ROOT}"/}/halt_${VERSION//[.-]/_}.gno"
	warn "this creates the proposal only — members vote and execute separately"
}

create_tag() {
	local message="${CHAIN} ${VERSION}

Released from chain/${CHAIN} at ${COMMIT:0:9}."
	[[ -z ${PREVIOUS} ]] || message+="
Previous release: ${PREVIOUS}."
	[[ -z ${HALT_HEIGHT} ]] || message+="
Coordinated upgrade at height ${HALT_HEIGHT}, halt_min_version ${VERSION}."

	git -C "${REPO_ROOT}" tag -a "${VERSION}" "${COMMIT}" -m "${message}"
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
	check_protocol_constants
	check_build_reports_tag
	printf '\n'

	classify

	[[ -z ${HALT_HEIGHT} ]] || emit_halt_proposal

	create_tag

	printf '\n'
	if [[ ${PUSH} -eq 1 ]]; then
		info "pushing"
		git -C "${REPO_ROOT}" push origin "refs/tags/${VERSION}:refs/tags/${VERSION}"
		ok "pushed ${VERSION} — release / chain-tag will build and attach the binaries"
	else
		info "dry run. To publish:"
		printf '\n    git push origin refs/tags/%s:refs/tags/%s\n\n' "${VERSION}" "${VERSION}"
		info "to undo the local tag: git tag -d ${VERSION}"
	fi
}

main "$@"
