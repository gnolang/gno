# Writing a scenario

A scenario is one claim about the chain, the oracle, or the harness, written so that the file fails when the claim
stops being true. Everything in it serves that claim: the cluster it asks for, the order of its lines, the reason
each wait waits, and the comment that says what a line proves. `testdata/tour.txtar` shows every
construct on this page; `docs/dialect.md` is the reference for the verbs and keys it uses.

## Shape

| Part | What goes there |
| --- | --- |
| header comment | the claim in one paragraph, then why it matters and what a failure would mean; no history |
| `# ---- ` sections | one per step of the argument, titled by what the step proves |
| script lines | the commands, each with a comment when its shape is not obvious |
| `-- cluster --` | the smallest chain the claim needs; comments allowed, keys documented in `docs/dialect.md` |
| package sections | the `.gno` files the script deploys, minimal, under the chain's domain |

Name the file for the claim, not the mechanism: `amnesiac_oracle.txtar` says what goes wrong; `restart_test.txtar`
would not. Put it in the directory for its subject, and add one line to the index in `README.md`.

## The cluster

Ask for the smallest chain that can exhibit the claim.

| Need | Ask for |
| --- | --- |
| a node and gnokey | `validators: 1` |
| a package that parks and an oracle that enables it | `code-submission-policy: inert` and nothing else; the oracle's key is the approver by default |
| an oracle that is not allowed to enable | `pkg-approver: user` |
| stopping a validator while the chain goes on | `validators: 4`; three tolerate no loss because a quorum is more than two thirds |
| a consensus or node setting the claim depends on | `block-max-gas`, `config.<key>` or `genesis.<key>`, and a line that reads the value back |

A setting that is declared and never read back is a hope. The tour reads its domain from `params/vm:p:chain_domain`,
its moniker from `$RPC_ADDR/status`, and its gas ceiling from the refusal that names it.

## One shot or wait

The dialect has two kinds of read, and choosing wrong produces a scenario that passes by luck or fails by timing.

| Situation | Read | Why |
| --- | --- | --- |
| the effect of a transaction this script just sent with `-broadcast=true`, read from the same node | one shot | `BroadcastTxCommit` returns once the transaction is in a committed block |
| the same effect read from another validator | `eventually` | a block commits on a quorum; the other node may not have applied that height yet |
| anything the oracle does | `eventually` | the enable is a second transaction the oracle sends when it next polls; no chain event announces it |
| a query that answers empty and exits 0 before the state arrives (`vm/qinertpaths`) | `eventually -stdout <regex>` | the exit status alone would end the wait on a node that is still behind |
| the oracle's status board | `eventually http_get <url> <regex>` | the board records an approval after the enable committed, so the package renders before the board says so |
| an absence: "nothing will enable this" | `sleep` then a one-shot read | an absence has no event; the comment states why the window is long enough (the tour uses two oracle poll ticks) |

`sleep` is never synchronization for something that will happen. `eventually` cannot be negated and cannot wrap a
testscript builtin.

## Negation

Every `!` line is followed by a `stderr` or `stdout` match that pins the reason, so the line cannot pass for a
different failure than the one the claim is about:

```
! gnokey maketx addpkg ... -gas-wanted 300000000 ...
stderr 'block-max-gas: 200000000'
```

A query for a parked or missing package prints `package "..." is not available` on stdout; pin that:

```
! gnokey query vm/qfile -data gno.land/r/probe/echo
stdout 'is not available'
```

Only `gnokey`, `http_get`, `gpao start` and `validator restart` are negatable. A `! validator restart N` asserts the
node cannot come back, and its error carries the node's stderr tail so the scenario can match the reason it died.

## Comments

A comment says what the line proves and why it is shaped the way it is, at a level the line itself cannot state:
why this is a wait and not a read, why this node and not another, why this number. When the reason lives in gno
code, name the function or file. Do not narrate what changed, do not describe neighbouring scenarios, and do not
restate the command. The existing scenarios set the register; read two before writing one.

## Determinism

- No exact heights, no wall-clock comparisons, no assertion that depends on which validator proposed.
- A limit or size the claim depends on is measured against the scenario's own transactions, and the comment says
  how.
- A validator that is stopped is one nothing earlier read from, so the stop cannot be blamed for a stale read.
- State the scenario disturbs is restored where a later step depends on it: a stopped validator restarts, a started
  oracle stops.

## Running while writing

```bash
cd misc/gnoe2e
go run . run -verbose testdata/oracle/<file>.txtar    # one scenario, every line echoed
go run . run testdata/oracle                          # the lane, quiet
make test-master SCENARIO_FLAGS='-v -timeout 30m -run TestScenarios/oracle/<name>'
```

Run a new scenario three times in a row before committing it. A scenario that fails once in three is a timing
assertion in disguise.

`make test-master` builds gnoland and gpao from `master` while running the working tree's scenarios. A scenario
written alongside a fix should be red there and green on the branch; the header says so in one bold line, and the
scenario lands with the fix rather than before it.

## Before committing

- [ ] the header states the claim and what a failure means
- [ ] the cluster is the smallest that exhibits it, and every declared setting is read back
- [ ] every wait says why it waits; every one-shot read is on the node that committed the write
- [ ] every `!` line is pinned by a `stderr` or `stdout` match
- [ ] no `sleep` stands in for a wait
- [ ] three consecutive passes, and the README index has the new line
