# PR6006: Handle queued internal messages before a timeout tick

## Context

`receiveRoutine` reads its inputs in one `select`, two of which can be ready at
the same instant:

```go
case mi = <-cs.internalMsgQueue:
	...
	cs.handleMsg(mi)
case ti := <-cs.timeoutTicker.Chan(): // tockChan:
	cs.wal.Write(ti)
	cs.handleTimeout(ti, rs)
```

`enterPropose` schedules the propose timeout and then calls `decideProposal`,
which signs the proposal and pushes it plus every block part onto
`internalMsgQueue`. Both arms are therefore ready together once the timeout
elapses, and Go picks between them uniformly at random.

When the tick wins, `handleTimeout` fires `EventTimeoutPropose` and calls
`enterPrevote` on the same goroutine, before the queue is drained. The proposer
prevotes nil for the block it signed a moment earlier, and its own proposal
lands one step too late.

Production nodes give the propose step three seconds, so a proposer only loses
that race when its consensus goroutine stalls for the whole step. Test nodes
give it 500ms, which a loaded CI runner reaches: `TestProposeValidBlock`,
`TestStateLockPOLSafety1`, `TestStartNextHeightCorrectly`, `TestStateLockNoPOL`
and `TestStateEnterProposeYesPrivValidator` all assert on a proposal the node
makes itself, and all five fail there.

The failure is worse than a lost round. Test subscriptions are unbuffered and
`FireEvent` delivers inline on `receiveRoutine`, so firing `EventTimeoutPropose`
at a test waiting on the proposal channel stops consensus dead until the 20s
`ensureTimeout` expires.

## Decision

Drain `internalMsgQueue` before letting a tick move the step:

```go
case ti := <-cs.timeoutTicker.Chan(): // tockChan:
	rs := cs.handleQueuedInternalMsgs()
	cs.wal.Write(ti)
	cs.handleTimeout(ti, rs)
```

`handleQueuedInternalMsgs` handles the messages already in the queue when the
tick was read, counted once at entry, and returns the round state they leave
behind. `handleTimeout` receives that state instead of the one snapshotted
before the `select`, so a proposal completed by the drain moves the step to
`RoundStepPrevote` and the tick is dropped by the guard already at the top of
`handleTimeout`. That snapshot had no other reader and is gone.

The count is what bounds the work. Handling one internal message can queue the
next: our own precommit completes a height, and under `SkipTimeoutCommit` the
next height's `enterPropose` runs inline from `addVote` and queues its proposal.
Draining until the queue reports empty would follow that chain for as long as
the node keeps producing, and `receiveRoutine` would stop reaching `cs.Quit()`.
A blocking receive is safe under the count because `receiveRoutine` is the only
reader of `internalMsgQueue`.

Messages on `internalMsgQueue` are the ones this node signed. Handling them
before a timeout that would overrule them is the order the state machine
intends: a node that has already produced its proposal or its vote should act on
it rather than time out waiting for it.

Each drained message keeps its `wal.WriteSync` and the tick is written after
them, so the WAL still records the order in which messages were handled and
replay is unchanged. The write and the `handleMsg` call move into
`handleInternalMsg`, shared with the `internalMsgQueue` arm of the `select`.

## Alternatives considered

**Raise `TimeoutPropose` in the affected tests.** The same tests wait on a
propose timeout in the rounds where the node is not the proposer, so the raise
is paid there in wall clock, about 4.5s per test, and the race is only made
rarer. Rejected.

**Tolerate the timeout in the test helpers.** An earlier revision of this branch
read the propose timeout while waiting for the proposal, which frees the routine
and ends the 20s hang. The prevote is still nil, so three of the four call sites
turned a hang into a `validatePrevote` failure two lines later rather than a
pass. Rejected as a fix, and dropped from the branch.

**Drain only for `RoundStepPropose` ticks.** Narrower, and it leaves the same
race on the vote path, where a queued precommit loses to the
`RoundStepPrevoteWait` tick. Rejected.

**Drain until the queue is empty rather than counting.** One line shorter, and
it hands `handleTimeout` a state the queue can no longer change. It also lets a
node that keeps queueing its own messages hold `receiveRoutine` inside the drain
across heights, so `cs.Quit()` stops being read and a `Stop` no longer lands
between messages. Rejected.

## Key files

| File | Role |
|------|------|
| `tm2/pkg/bft/consensus/state.go` | `receiveRoutine`, `handleQueuedInternalMsgs`, `handleInternalMsg` |
| `tm2/pkg/bft/consensus/state_test.go` | `TestStateEnterProposeNoTimeoutOnOwnProposal` |

## Consequences

- A proposer no longer prevotes nil for its own proposal because the propose
  step expired while the proposal sat in the queue.
- The five consensus tests above stop depending on the outcome of that race.
  With `TimeoutPropose` forced to a microsecond, so that the tick is ready every
  time:

  | test | failures before | failures after |
  |---|---|---|
  | `TestProposeValidBlock` | 4 of 6 | 0 of 6 |
  | `TestStateLockPOLSafety1` | 2 of 6 | 0 of 6 |
  | `TestStartNextHeightCorrectly` | 5 of 6 | 0 of 6 |
  | `TestStateLockNoPOL` | 4 of 6 | 0 of 6 |
  | `TestStateEnterProposeYesPrivValidator` | 2 of 6 | 0 of 6 |

- `TestStateEnterProposeNoTimeoutOnOwnProposal` takes the proposal of eight
  consecutive heights with the same microsecond timeout, and fails on the first
  `EventTimeoutPropose` it sees. It fails 6 times out of 6 without the change.
- A tick now waits for the messages this node queued for itself. The queue holds
  work `receiveRoutine` would run next in any case, so the tick is delayed by
  the handling of messages already signed, not by new work. A drain that
  finishes a height writes its `EndHeightMessage` before the tick, so replay can
  read a tick after a height boundary; `handleTimeout` drops it on the height
  comparison, in replay through `readReplayMessage` as at runtime.
