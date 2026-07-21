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
