#!/usr/bin/env bash
# promote-release.sh — make a final release the one operators land on.
#
# Every release is published held: release / chain-tag publishes it as a
# GitHub pre-release, and release / docker tags its images by version only. A
# final release is out days before the chain switches to it, and until then
# the release the chain runs keeps GitHub's "Latest", misc/install.sh's default
# and the CLI images' `latest`. This script moves all three to the final once
# the switch is set: for a coordinated upgrade, once the halt proposal has
# executed; for a rolling one, on the agreed day.
#
# A promotion only moves forward from the highest full release, which is not
# always the highest tag: a patch to the running line can ship while the next
# MINOR is held for its halt. A MINOR is refused until mainnet's halt proposal,
# read from mainnet's RPC, names it; a PATCH needs no halt, and a MAJOR, a new
# network, is never promoted here.
#
# Usage:
#   misc/release/promote-release.sh <version> [--push]
#
# Options:
#   --push   promote the release (otherwise: dry run, every check, nothing changed)
#
# Needs curl, jq, sha256sum or shasum, and gh logged in with the write:packages
# scope.
#
# Example:
#   misc/release/promote-release.sh v1.7.0 --push

set -euo pipefail

# Regex bracket ranges follow the locale's collation; in the C locale the shape
# check matches the node's version parser.
export LC_ALL=C

readonly REPO="gnolang/gno"
readonly REGISTRY="ghcr.io"
# Where mainnet's scheduled halt is read; overridable for a test.
readonly MAINNET_RPC="${MAINNET_RPC:-https://rpc.gno.land}"

# The CLI tools. Never gnoland: a node image that moves by itself breaks a
# coordinated upgrade (RELEASING.md, "Container images").
# gno.land/pkg/gnoland.TestReleasesArePublishedHeld reads this line and checks
# every image is one release / docker builds, so keep it on one line.
readonly LATEST_IMAGES=(gno gnokey gnodev gnoweb)

# The shape cut-release.sh accepts; gno.land/pkg/gnoland.TestReleaseToolingMatchesTheParser
# fails if the two drift, so keep it on one line.
readonly VERSION_RE='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*)?$'

# Every manifest shape release / docker pushes. The registry serves the stored
# bytes under one of them, and the same type goes back with the bytes.
readonly MANIFEST_TYPES=(
  application/vnd.oci.image.index.v1+json
  application/vnd.docker.distribution.manifest.list.v2+json
  application/vnd.oci.image.manifest.v1+json
  application/vnd.docker.distribution.manifest.v2+json
)
MANIFEST_ACCEPT="$(
  IFS=,
  printf '%s' "${MANIFEST_TYPES[*]}"
)"
readonly MANIFEST_ACCEPT

VERSION=""
PUSH=0
# Every full release of REPO (neither a draft nor a pre-release), one tag per line.
RELEASES=""
# The highest final full release: the one a promotion moves forward from.
LATEST_NOW=""
# The GitHub account the registry tokens are asked for.
LOGIN=""
SCRATCH=""
# Set once the run has ended on purpose. bash 3.2 ends a script aborted by
# set -u with the EXIT trap's own status, so the trap fails any other ending.
DONE=0
trap '[[ -z ${SCRATCH} ]] || rm -rf "${SCRATCH}"; [[ ${DONE} -eq 1 ]] || exit 1' EXIT

die() {
  printf '\033[31merror:\033[0m %s\n' "$*" >&2
  exit 1
}
info() { printf '\033[36m==>\033[0m %s\n' "$*"; }
ok() { printf '  \033[32m✓\033[0m %s\n' "$*"; }
# Up to the first blank line, so adding an option to the header above cannot
# silently truncate --help.
usage() { sed -n '2,/^$/p' "${BASH_SOURCE[0]}" | sed 's|^# \{0,1\}||'; }

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
    -h | --help)
      usage
      DONE=1
      exit 0
      ;;
    --push)
      PUSH=1
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

check_tools() {
  local tool
  # No wget fallback: the registry calls rely on curl's config on stdin (-K -),
  # header dumps (-D), status codes (-w) and PUT, which busybox wget lacks.
  for tool in gh curl jq; do
    command -v "${tool}" >/dev/null 2>&1 || die "needs ${tool}"
  done
  # GNU coreutils has sha256sum, macOS has shasum.
  if command -v sha256sum >/dev/null 2>&1; then
    sha256() { sha256sum "$@"; }
  elif command -v shasum >/dev/null 2>&1; then
    sha256() { shasum -a 256 "$@"; }
  else
    die "needs sha256sum or shasum"
  fi
}

# ------------------------------------------------------------------- checks

# A release candidate is a rehearsal, never the release operators land on.
check_final() {
  [[ ${VERSION} =~ ${VERSION_RE} && ${VERSION} != *-* ]] ||
    die "only a final vMAJOR.MINOR.PATCH release is promoted (got ${VERSION}).
       A release candidate stays a pre-release."
  ok "final release: ${VERSION}"
}

check_login() {
  LOGIN="$(gh api user --jq .login)" || die "gh is not logged in (gh auth login)"
  ok "gh is logged in as ${LOGIN}"
}

# release / chain-tag publishes the release once every binary is attached, so a
# draft, or no release at all, means its run has not finished.
check_published() {
  local draft
  draft="$(gh release view "${VERSION}" --repo "${REPO}" --json isDraft --jq .isDraft)" ||
    die "no release ${VERSION} on ${REPO}: was the tag pushed, and has release / chain-tag run?"
  [[ ${draft} == false ]] || die "${VERSION} is still a draft: release / chain-tag has not finished"
  ok "release ${VERSION} is published"
}

read_releases() {
  RELEASES="$(gh api --paginate "repos/${REPO}/releases" \
    --jq '.[] | select(.draft == false and .prerelease == false) | .tag_name')" ||
    die "cannot list the releases of ${REPO}"
}

# The highest final release in RELEASES. Not GitHub's "Latest" marker: a
# chain/<name> genesis record is a full release created by hand, and can hold it.
pick_latest() {
  local tag finals=""
  while read -r tag; do
    if [[ ${tag} =~ ${VERSION_RE} && ${tag} != *-* ]]; then
      finals+="${tag}"$'\n'
    fi
  done <<<"${RELEASES}"
  [[ -n ${finals} ]] || die "${REPO} has no full release to move forward from"
  LATEST_NOW="$(printf '%s' "${finals}" | sort -V | tail -n 1)" || die "sort -V failed"
}

# The latest release is the one the chain runs, so a promotion never goes back
# to an older line, whichever release was cut last.
check_moves_forward() {
  local newest
  newest="$(printf '%s\n%s\n' "${LATEST_NOW}" "${VERSION}" | sort -V | tail -n 1)" || die "sort -V failed"
  if [[ ${LATEST_NOW} == "${VERSION}" ]]; then
    ok "${VERSION} is the latest release already"
  elif [[ ${newest} == "${VERSION}" ]]; then
    ok "moves forward: ${LATEST_NOW} → ${VERSION}"
  else
    die "${LATEST_NOW} is the latest release, and ${VERSION} is older: a promotion only moves forward"
  fi
}

# Mainnet's halt_min_version, read from its RPC. An answer that is not a query
# result is refused, never taken for an empty value.
halt_min_version() {
  local body answer value
  body='{"jsonrpc":"2.0","id":1,"method":"abci_query","params":{"path":"params/node:p:halt_min_version","data":""}}'
  answer="$(curl -q -sS --connect-timeout 10 --max-time 30 -X POST -H 'Content-Type: application/json' -d "${body}" "${MAINNET_RPC}")" ||
    die "cannot ask ${MAINNET_RPC} for mainnet's halt_min_version"
  value="$(jq -er '.result.response.ResponseBase | if .Error != null then error("query error") else (.Data // error("no data")) | @base64d | fromjson end' <<<"${answer}" 2>/dev/null)" ||
    die "${MAINNET_RPC} gave no halt_min_version: ${answer:0:200}"
  printf '%s\n' "${value}"
}

# A MINOR is a coordinated upgrade: it becomes the release operators land on
# only once mainnet's halt proposal names it. A PATCH switches without a halt,
# the release the chain runs needs none, and a MAJOR is a new network, which
# mainnet never schedules.
check_scheduled() {
  local now_major now_minor major minor scheduled
  IFS=. read -r now_major now_minor _ <<<"${LATEST_NOW#v}"
  IFS=. read -r major minor _ <<<"${VERSION#v}"
  [[ ${major} == "${now_major}" ]] ||
    die "${VERSION} is a MAJOR from ${LATEST_NOW}: a MAJOR is a new network, and promote-release.sh only promotes along mainnet's line"
  if [[ ${VERSION} == "${LATEST_NOW}" || ${minor} == "${now_minor}" ]]; then
    ok "no halt to check: ${VERSION} is not a new MINOR"
    return
  fi
  scheduled="$(halt_min_version)" || exit 1
  [[ ${scheduled} == "${VERSION}" ]] ||
    die "mainnet has not scheduled ${VERSION}: its halt_min_version is '${scheduled}'. A MINOR is promoted once its halt proposal has executed."
  ok "mainnet's halt_min_version is ${VERSION}"
}

# ------------------------------------------------------------------- images

# curl for the registry: without ~/.curlrc, where a `verbose` line would print
# the Authorization header, and failing on a stalled connection.
registry() { curl -q -sS --connect-timeout 10 --max-time 120 "$@"; }

# The value of one header in a curl -D dump, parameters included.
header() {
  tr -d '\r' <"$2" | awk -v k="$1" 'tolower($1) == k ":" { sub(/^[^:]*:[ \t]*/, ""); v = $0 } END { print v }'
}

# Writes the curl config holding a bearer token for one image's repository into
# the private scratch directory. Neither token reaches an argument list or an
# xtrace line: gh's goes to curl on stdin.
registry_auth() {
  local image=$1 actions=$2 gh_token token xtrace=0
  [[ $- != *x* ]] || {
    xtrace=1
    set +x
  }
  gh_token="$(gh auth token)" || die "gh has no token (gh auth login)"
  token="$(printf 'user = "%s:%s"\n' "${LOGIN}" "${gh_token}" |
    registry -f -K - "https://${REGISTRY}/token?scope=repository:${REPO}/${image}:${actions}" |
    jq -er .token)" ||
    die "cannot get a ${actions} token for ${REGISTRY}/${REPO}/${image}"
  printf 'header = "Authorization: Bearer %s"\n' "${token}" >"${SCRATCH}/${image}/auth"
  [[ ${xtrace} -eq 0 ]] || set -x
}

# Reads <image>:<version>'s manifest into the scratch directory, its bytes as
# stored, and checks them against their digest. With --push it asks a token
# that may write, but ghcr.io checks the permission at the write itself, so a
# missing one shows at the first write.
read_image() {
  local image=$1 dir="${SCRATCH}/$1" actions=pull digest type sum known
  mkdir "${dir}"
  [[ ${PUSH} -eq 0 ]] || actions=pull,push
  registry_auth "${image}" "${actions}"
  registry -f -K "${dir}/auth" -H "Accept: ${MANIFEST_ACCEPT}" -D "${dir}/headers" -o "${dir}/manifest" \
    "https://${REGISTRY}/v2/${REPO}/${image}/manifests/${VERSION}" ||
    die "cannot read ${REGISTRY}/${REPO}/${image}:${VERSION}: has release / docker finished?"
  digest="$(header docker-content-digest "${dir}/headers")"
  sum="$(sha256 "${dir}/manifest")" || die "cannot hash ${image}:${VERSION}'s manifest"
  [[ "sha256:${sum%% *}" == "${digest}" ]] ||
    die "${REGISTRY} named ${image}:${VERSION} ${digest:-by no digest}, and its bytes hash to sha256:${sum%% *}"
  type="$(header content-type "${dir}/headers")"
  for known in "${MANIFEST_TYPES[@]}"; do
    [[ ${type} != "${known}" ]] || break
  done
  [[ ${type} == "${known}" ]] || die "${REGISTRY} served ${image}:${VERSION} as ${type:-no media type}, not a manifest type release / docker pushes"
  printf '%s\n' "${digest}" >"${dir}/digest"
  printf '%s\n' "${type}" >"${dir}/type"
  ok "${image}:${VERSION} is ${digest} (${type})"
}

# The digest <image>:latest answers with, or nothing.
latest_digest() {
  local dir="${SCRATCH}/$1"
  if registry -f -K "${dir}/auth" -I -H "Accept: $(cat "${dir}/type")" -D "${dir}/latest" -o /dev/null \
    "https://${REGISTRY}/v2/${REPO}/$1/manifests/latest"; then
    header docker-content-digest "${dir}/latest"
  fi
}

# Points <image>:latest at <image>:<version> by putting the version's manifest
# under latest byte for byte: a re-serialized one would give latest a digest no
# release names. The digest latest answers with afterwards proves the bytes.
move_latest() {
  local image=$1 dir="${SCRATCH}/$1" digest code
  digest="$(cat "${dir}/digest")"
  if [[ "$(latest_digest "${image}")" == "${digest}" ]]; then
    ok "${image}:latest = ${image}:${VERSION} already"
    return
  fi
  code="$(registry -K "${dir}/auth" -X PUT -H "Content-Type: $(cat "${dir}/type")" \
    --data-binary "@${dir}/manifest" -o "${dir}/put" -w '%{http_code}' \
    "https://${REGISTRY}/v2/${REPO}/${image}/manifests/latest")" || code="none"
  [[ ${code} == 201 ]] ||
    die "${REGISTRY} refused ${image}:latest (HTTP ${code}): $(cat "${dir}/put" 2>/dev/null). Run the script again once fixed."
  [[ "$(latest_digest "${image}")" == "${digest}" ]] ||
    die "${image}:latest does not answer with ${VERSION}'s ${digest}. Run the script again once fixed."
  ok "${image}:latest = ${image}:${VERSION}"
}

# --------------------------------------------------------------------- main

main() {
  parse_args "$@"
  check_tools

  info "checks"
  check_final
  check_login
  check_published
  read_releases
  pick_latest
  check_moves_forward
  check_scheduled

  # Every image is read and checked before anything is written, so a missing
  # or corrupt one leaves nothing half-promoted.
  info "images"
  SCRATCH="$(mktemp -d)"
  local image latest
  for image in "${LATEST_IMAGES[@]}"; do
    read_image "${image}"
  done

  if [[ ${PUSH} -eq 0 ]]; then
    printf '\n'
    info "dry run — nothing was changed. With --push, ${VERSION} becomes Latest and"
    info "latest moves to it on ${LATEST_IMAGES[*]}"
    return
  fi

  # Images first, the release last: a run stopped halfway leaves the release a
  # pre-release, and running the script again finishes it.
  printf '\n'
  info "promoting"
  for image in "${LATEST_IMAGES[@]}"; do
    move_latest "${image}"
  done
  gh release edit "${VERSION}" --repo "${REPO}" --prerelease=false --latest >/dev/null ||
    die "latest points at ${VERSION} on every image, but the release edit failed. Run the script again once fixed."
  latest="$(gh api "repos/${REPO}/releases/latest" --jq .tag_name)" || die "cannot read GitHub's Latest back"
  [[ ${latest} == "${VERSION}" ]] || die "GitHub ranks ${latest} Latest, not ${VERSION}"
  ok "${VERSION} is a full release, and Latest"
}

main "$@"
DONE=1
