# gno-fixes — internal security-fix staging

> **This is a private staging repo.** It exists to develop and review security
> fixes for [gnolang/gno](https://github.com/gnolang/gno) *before* they are
> ported upstream, in anonymized batches, so that individual commits are not
> obviously tied to a specific vulnerability report or advisory.
>
> Treat everything here as sensitive until it has landed upstream.

## Branch model

```
upstream/master  ← the public gnolang/gno:master. Not mirrored here; add it as
   (not here)      a remote and branch off it. This repo has no `master`.
develop*         ← default branch: upstream master + all in-flight security
                   fixes + this guide. PRs land here.
                     │
                     ├── dev/<user>/<slug>   ─PR→ develop
                     └── dev/<user>/<slug>   ─PR→ develop
(* repo default branch)
```

- **Fix branches are based on the public `gnolang/gno:master`**, not on any
  branch of this repo, so a fix stays a clean commit against current upstream
  and is cherry-pickable at port time. There used to be an auto-synced `master`
  mirror here; it was removed (see [Notes](#notes)) — fetch upstream directly.
- **`develop`** is where work is integrated and reviewed. It is the default
  branch, so PRs target it automatically. Keep it current by merging
  `upstream/master` into it.

## Working on a fix

1. Branch off the public upstream `master`, so your fix is a clean commit
   against current upstream and stays cherry-pickable at port time:
   ```
   git remote add upstream https://github.com/gnolang/gno.git   # once
   git fetch upstream master
   git switch -c dev/<user>/<short-slug> upstream/master
   ```
2. Name the branch `dev/<user>/<slug>` — **never** put an advisory ID, bounty
   reference, or "CVE"/"vuln" wording in the branch name, commit messages, or
   PR title. Describe the fix by what it does, not what it patches.
3. Include an **ADR** (`gnovm/adr/`, `gno.land/adr/`, or `tm2/adr/` per scope)
   and regression tests. See the repo's root `AGENTS.md` / `CONTRIBUTING.md`.
4. Open the PR against **`develop`** (the default). Get it reviewed and merged.

## Porting upstream (maintainers)

Fixes are ported to `gnolang/gno` in batches to obscure the 1:1 mapping to
reports. **Each batch goes upstream as a single squashed commit** with an
ordinary-looking message and every batch contributor credited as a co-author:

1. Start from a clean checkout of `upstream/master`, on a new branch.
2. Bring in the fixes for this batch — cherry-pick the fix commits (by SHA) or
   `git merge --squash` their branches — but **never** include merge commits or
   anything under `.github/`; this guide is internal-only and must never leave
   this repo.
3. Collapse the whole batch into **one commit** (e.g. `git reset --soft
   upstream/master && git commit`).
4. Write a **plausible, non-security commit message**: a normal `feat:` /
   `fix:` / `refactor:` / `perf:` that matches what the touched code does. Do
   **not** hint at a vulnerability, DoS, advisory, or "hardening" — frame it as
   routine maintenance (e.g. bounding/streaming for perf, an API tidy-up).
5. Credit everyone whose work is in the batch with `Co-authored-by:` trailers
   (one per contributor, using their GitHub-registered name/email), so
   attribution is preserved without the commit reading as a security fix:
   ```
   perf(gnovm): stream query value export to bound response size

   Co-authored-by: Alice <alice@example.com>
   Co-authored-by: Bob <bob@users.noreply.github.com>
   ```
6. Rename/drop any internal-only giveaways before pushing — e.g. `prxxxx_*`
   ADR filenames, references to this repo, or report IDs.
7. Open one ordinary-looking upstream PR. Space batches out so no single PR
   reads as "the fix for X".

## Notes

- CI runs the inherited gno `ci-*` test workflows on PRs to `develop`, so fixes
  are tested against real upstream code. Their upstream `push: branches:
  [master]` triggers were repointed at `develop`, so merging a PR also runs the
  full suite (the `ci-dir-*` push triggers have no path filter, by design
  upstream). The **upstream-only** workflows
  (deploys, releases, CodeQL, FOSSA, the Discord/GitHub bots, and PR-hygiene
  automation) were removed from `develop` — they need secrets, deploy targets,
  or GitHub Advanced Security that this private fork doesn't have, and only
  produced noise. Re-add anything here on `develop` if you want it.
- The `ci-*` jobs occasionally fail at the checkout step with "Repository not
  found" — a transient private-repo hiccup, not a real failure. Just re-run the
  job (`gh run rerun <id> --failed`).
- **Why there is no `master` mirror.** A scheduled Action used to fast-forward a
  local `master` from `gnolang/gno:master`, but it can't work with
  `GITHUB_TOKEN`: GitHub rejects any GitHub App push that touches
  `.github/workflows/**`, so the sync broke on every upstream commit that edited
  a workflow. The workarounds all cost more than the mirror was worth — a PAT or
  GitHub App token *can* push workflow files, but it is a real actor, so its
  pushes trigger workflow runs, and `master` carries the full upstream set
  (`ci-dir-*`, `ci-e2e`, `deploy-pages`, `release-*` all fire unfiltered on push
  to `master`). That is a full CI plus deploy/release run every hour on billed
  private-repo minutes, and the shared `ci-*` entries can't be disabled without
  also disabling PR CI on `develop`. Since nothing ever PR'd against `master`,
  it only served as a branching base — and the public repo already is one.
