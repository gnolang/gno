# gh-report

Dense GitHub report of open issues and PRs that deserve attention.

## Usage

    make fetch    # one GraphQL call per repo in repos.txt
    make report   # render from data/ to stdout

## Output modes

    go run ./cmd/gh-report           # markdown (default)
    go run ./cmd/gh-report --ansi    # ANSI colors
    go run ./cmd/gh-report --json    # JSON

See `specs/2026-05-20-gh-report-design.md` for the full design.

## Prior art

Samourai run a [weekly-report skill](https://github.com/samouraiworld/gno-agent-workspace/blob/main/skills/weekly-report.md)
over the same repositories, driven by
[`weekly-report.sh`](https://github.com/samouraiworld/gno-agent-workspace/blob/main/scripts/weekly-report.sh)
and [`parse-context.sh`](https://github.com/samouraiworld/gno-agent-workspace/blob/main/scripts/parse-context.sh)
(pointed out by @davd-gzl on #5703; [sample output](https://github.com/samouraiworld/gno-agent-workspace/blob/main/reports/weekly/2026-05-17/report.md)).

The two solve different halves of the problem and are worth keeping distinct:

- Theirs is prose-first and agent-driven — a shell pipeline gathers context and
  a model writes the narrative. Good for a weekly digest a human reads once.
- This one is deterministic and diff-able: one GraphQL query per repo, then a
  pure renderer with golden tests, no model in the loop. The same input always
  produces the same bytes, which is what makes it usable as a triage list you
  re-run daily and compare against yesterday's.

If the two converge, the natural split is this tool as the data layer (its
`--json` mode is meant for exactly that) feeding a prose pass like theirs,
rather than either one growing the other's half.
