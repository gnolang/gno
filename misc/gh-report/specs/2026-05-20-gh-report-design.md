# gh-report — design

Date: 2026-05-20
Branch: `gh-report`
Status: draft, awaiting user review

## Goal

A maintainer-focused tool that aggregates open GitHub issues and PRs across one
or more repos (default: `gnolang/gno`) updated in the last 30 days, classifies
each entry into one or more attention buckets, and emits a dense, one-line-per-
entry report. Designed to surface what needs review, what is stuck, and what is
ready to merge, with minimal noise.

Two-phase by design:

1. **Fetch** — shell script using `gh api graphql` writes raw JSON per repo.
2. **Report** — Go binary reads JSON, classifies, renders.

Re-running the report after tuning thresholds costs nothing (no API hit).

## Stack and layout

Hybrid: shell for fetch (one GraphQL call per repo), Go for report.

```
misc/gh-report/
  README.md
  Makefile              # fetch, report, all, test, install, clean
  repos.txt             # owner/repo per line; default: gnolang/gno
  query.graphql         # single rich query reused per repo
  fetch.sh              # loop repos.txt, write data/<owner>--<repo>.json
  data/                 # gitignored, raw GraphQL responses per repo
  specs/                # this file
  testdata/             # JSON fixtures for classify_test.go
  cmd/gh-report/
    main.go             # flags, load data, dispatch render
    classify.go         # constants + classification rules
    render.go           # markdown / ANSI / JSON renderers
    types.go            # Entry, Bucket, Report
  go.mod
```

## Fetch phase

`fetch.sh`:

- `set -euo pipefail`.
- Reads `repos.txt`, skips blank lines and `#` comments.
- For each `owner/name`, runs:
  ```
  gh api graphql -F owner=$OWNER -F name=$NAME -f query="$(cat query.graphql)" \
    > data/$OWNER--$NAME.json
  ```
- Paginates until the oldest returned `updatedAt` is older than 30 days (early
  exit so we do not fetch the whole project history).
- Errors:
  - `gh` missing or unauthenticated: clear stderr, exit 1.
  - Rate limit hit: print `X-RateLimit-Reset`, exit 1.
  - Repo 404: warn and skip that repo, continue with the next.
  - Network failure: exit 1 with readable stderr.

`query.graphql` retrieves, per entry:

- `number, title, url, createdAt, updatedAt`
- `author { login, ... on User { createdAt }, ... on Bot { __typename } }` plus
  `authorAssociation` (used for first-contribution detection)
- `assignees(first:10) { nodes { login } }`
- `labels(first:20) { nodes { name } }`
- `reactions { totalCount }`
- `comments(last:5) { totalCount, nodes { author { login }, createdAt, body } }`
- `timelineItems(last:10, ...)` for PRs (review requested events)

PR-only fields:

- `reviews(last:50) { nodes { state, author { login }, submittedAt } }`
- `reviewRequests(first:20) { nodes { requestedReviewer { ... } } }`
- `commits(last:1) { nodes { commit { statusCheckRollup { state } } } }`
- `mergeable`
- `isDraft`

## Classification

Multi-tag: an entry can appear in several sections. Constants live at the top of
`classify.go` and are tuned in code, not via flags.

```go
const (
    WINDOW_DAYS         = 30  // aggregation window
    STALE_DAYS          = 14
    STUCK_OPEN_DAYS     = 30
    STUCK_NO_UPDATE_DAYS = 7
    HOT_RECENT_DAYS     = 7
    HOT_COMMENTS        = 5
    HOT_REACTIONS       = 3
    NEW_CONTRIB_DAYS    = 90  // account age threshold
)

var (
    JAEKWON     = "jaekwon"
    MOUL        = "moul"
    OTHER_CORE  = []string{"zivkovicmilos", "thehowl", "leohhhn"} // tune over time
    EXCLUDE_LABELS = []string{"wontfix", "duplicate", "invalid"}
)
```

Sections (in report order):

| # | Section | Rule |
|---|---------|------|
| 1 | Hot | `recent_comments >= HOT_COMMENTS` OR `reactions >= HOT_REACTIONS` over last `HOT_RECENT_DAYS` |
| 2 | Ready to merge | PR + at least one `APPROVED` review + statusCheckRollup `SUCCESS` + no reviewer whose latest review is `CHANGES_REQUESTED` + `mergeable == MERGEABLE` + not draft |
| 3 | Depends on @jaekwon | assignee == jaekwon OR requested reviewer == jaekwon OR `@jaekwon` substring in body of any of the last 5 comments |
| 4 | Depends on @moul | same |
| 5 | Depends on other core | reviewer/assignee/mention matches any handle in `OTHER_CORE`. Line includes the matched handle. |
| 6 | From new contributors | `authorAssociation` in `{FIRST_TIMER, FIRST_TIME_CONTRIBUTOR, NONE}` OR author account age < `NEW_CONTRIB_DAYS` |
| 7 | Stuck | open > `STUCK_OPEN_DAYS` AND a review was requested AND no update for `STUCK_NO_UPDATE_DAYS` |
| 8 | Stale | no update for `STALE_DAYS` AND not in sections 1-6 |

### Edge cases

- PR with `mergeable == UNKNOWN`: not eligible for "Ready to merge" (wait for
  GitHub to compute).
- Draft PRs: excluded from all sections except Hot.
- Entries with any label in `EXCLUDE_LABELS`: excluded entirely.
- Bot-driven updates (`dependabot[bot]`, `mergify[bot]`, etc.): `updatedAt` is
  not counted as human activity for Hot/Stuck. Use comment authors and review
  authors instead.
- Bot authors: excluded from "From new contributors".

## CLI

```
gh-report                       # markdown to stdout (default)
gh-report --ansi                # plain text with ANSI colors for terminal
gh-report --json                # structured JSON
gh-report --data ./data         # override data dir (default: ./data)
gh-report --repo gnolang/gno    # restrict to one repo
```

## Rendering

Three modes. Section order matches the classification table. Empty sections are
omitted. Each section header includes a count. An entry can appear in multiple
sections (dedup is not desired).

Line format (compact, ~80-100 chars):

```
- #1234 [PR/3d] title truncated to ~70 chars (author, +5c)
```

Fields: number, type (`issue` or `PR`), days since update, truncated title,
author, comment count.

Markdown mode emits `[#1234](url)` so the number is clickable. ANSI mode keeps
plain `#1234`. The "Depends on other core" section adds the matched reviewer
inside the metadata bracket: `[PR/3d/@thehowl]`.

Example output:

```markdown
## Hot (3)
- [#5612](https://github.com/gnolang/gno/pull/5612) [PR/2d] feat(gnoweb): accept gno.land URLs (moul, +12c)
- [#5293](https://github.com/gnolang/gno/issues/5293) [issue/5d] gnovm: storage gas overrun (jaekwon, +8c)

## Ready to merge (1)
- [#5685](https://github.com/gnolang/gno/pull/5685) [PR/0d] fix(ci): persist credentials (moul, +2c)

## Depends on @jaekwon (4)
...

## Depends on other core (2)
- [#5411](https://github.com/gnolang/gno/pull/5411) [PR/3d/@thehowl] refactor: ...
```

ANSI mode: headers bold cyan, numbers yellow, dates older than `STALE_DAYS` red.

JSON mode:

```json
{
  "generated_at": "2026-05-20T...",
  "window_days": 30,
  "sections": [
    {"name": "Hot", "count": 3, "entries": [{"number": 5612, "type": "PR", ...}]}
  ]
}
```

## Errors and edge cases in the report

- `data/` empty or missing: print "no data, run `make fetch` first", exit 0.
- Malformed JSON in one file: log the offending file, skip it, continue.
- Repo in `data/` not listed in `repos.txt`: include it anyway. Data on disk
  wins.
- All sections empty: print "no items in window".

## Testing

- `classify_test.go`: table-driven, fixtures in `testdata/`. About 10-15 cases
  covering each rule plus the edge cases above. One fixture file per scenario.
- `render_test.go`: golden tests for markdown and JSON. No ANSI test (visual).
- `fetch.sh`: not tested automatically (touches the API). Validated by hand on
  first run.
- Not tested: end-to-end integration (covered by usage), GraphQL query syntax
  (validated by GitHub at runtime).

`go test ./...` from `misc/gh-report/`.

## Out of scope

- Closed/merged entries (could be added later as a separate report).
- LLM-based classification.
- Webhook or scheduled runs (just a Makefile, run when needed).
- Configurable thresholds via flags (tune the Go constants).
- Cross-repo deduplication (one entry per repo, never duplicated across repos
  by design since numbers are repo-scoped).
