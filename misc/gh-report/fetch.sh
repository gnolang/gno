#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
DATA_DIR="${DATA_DIR:-$ROOT/data}"
REPOS_FILE="${REPOS_FILE:-$ROOT/repos.txt}"
QUERY_FILE="${QUERY_FILE:-$ROOT/query.graphql}"
WINDOW_DAYS="${WINDOW_DAYS:-30}"

mkdir -p "$DATA_DIR"

if ! command -v gh >/dev/null 2>&1; then
  echo "gh-report: 'gh' CLI not installed" >&2
  exit 1
fi
if ! gh auth status >/dev/null 2>&1; then
  echo "gh-report: 'gh' not authenticated (run: gh auth login)" >&2
  exit 1
fi

# cutoff: now - WINDOW_DAYS, as RFC3339
cutoff=$(date -u -d "${WINDOW_DAYS} days ago" +%Y-%m-%dT%H:%M:%SZ 2>/dev/null \
  || gdate -u -d "${WINDOW_DAYS} days ago" +%Y-%m-%dT%H:%M:%SZ)

while IFS= read -r line; do
  # skip blanks and comments
  [[ -z "${line// }" || "$line" =~ ^# ]] && continue
  repo="$line"
  owner="${repo%%/*}"
  name="${repo##*/}"
  out="$DATA_DIR/${owner}--${name}.json"

  echo ">> fetching $repo (cutoff $cutoff)" >&2

  query="$(cat "$QUERY_FILE")"
  # Single page for now. Pagination can be added if a repo has >100 open issues
  # or PRs updated within the window.
  if ! gh api graphql \
        -F owner="$owner" -F name="$name" \
        -f query="$query" > "$out.tmp"; then
    rc=$?
    if [[ $rc -ne 0 ]]; then
      echo "gh-report: failed to fetch $repo (rc=$rc), skipping" >&2
      rm -f "$out.tmp"
      continue
    fi
  fi

  # Detect rate-limit errors in body.
  if grep -q '"type":"RATE_LIMITED"' "$out.tmp"; then
    echo "gh-report: rate-limited fetching $repo" >&2
    rm -f "$out.tmp"
    continue
  fi

  mv "$out.tmp" "$out"
  echo ">> wrote $out" >&2
done < "$REPOS_FILE"
