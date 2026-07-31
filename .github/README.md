# gno-fixes — internal security-fix staging

> **This is a private staging repo.** It exists to develop and review security
> fixes for [gnolang/gno](https://github.com/gnolang/gno) *before* they are
> ported upstream, in anonymized batches, so that individual commits are not
> obviously tied to a specific vulnerability report or advisory.
>
> Treat everything here as sensitive until it has landed upstream.

## Branch model

```
master    ← pristine, auto-synced mirror of gnolang/gno:master (a bot keeps it
            fast-forwarded; DO NOT push to it or open PRs against it)
develop*  ← default branch: master + all in-flight security fixes + this guide
            + the sync workflow. PRs land here.
              │
              ├── dev/<user>/<slug>   ─PR→ develop
              └── dev/<user>/<slug>   ─PR→ develop
(* repo default branch)
```

- **`master`** is a hands-off mirror. A scheduled GitHub Action
  (`.github/workflows/sync-upstream.yml`, runs from `develop`) fast-forwards it
  from `gnolang/gno:master`. Because nothing is ever committed to it directly,
  the sync never conflicts, and it is always a clean base to build port PRs
  from.
- **`develop`** is where work is integrated and reviewed. It is the default
  branch, so PRs target it automatically.

## Working on a fix

1. Branch off **`master`** (the pristine mirror), so your fix is a clean commit
   against current upstream and stays cherry-pickable at port time:
   ```
   git fetch origin
   git switch -c dev/<user>/<short-slug> origin/master
   ```
2. Name the branch `dev/<user>/<slug>` — **never** put an advisory ID, bounty
   reference, or "CVE"/"vuln" wording in the branch name, commit messages, or
   PR title. Describe the fix by what it does, not what it patches.
3. Include an **ADR** (`gnovm/adr/`, `gno.land/adr/`, or `tm2/adr/` per scope)
   and regression tests. See the repo's root `AGENTS.md` / `CONTRIBUTING.md`.
4. Open the PR against **`develop`** (the default). Get it reviewed and merged.

## Porting upstream (maintainers)

Fixes are ported to `gnolang/gno` in batches to obscure the 1:1 mapping to
reports:

1. Start from a clean checkout of `origin/master` (which equals upstream).
2. Cherry-pick the individual fix commits you want to ship this batch (by SHA),
   **not** merge commits and **not** anything under `.github/` — this guide and
   the sync workflow are internal-only and must never leave this repo.
3. Reword/split/reorder as needed to anonymize, then open normal-looking PRs
   upstream. Combine or space them out so no single PR reads as "the fix for X".

## Notes

- CI (the inherited gno workflows) runs on PRs to `develop`, so fixes are tested
  against real upstream code.
- If the sync Action ever fails, it usually means someone pushed to `master` by
  mistake — reset `master` back to `gnolang/gno:master` and move the change onto
  a `dev/...` branch.
