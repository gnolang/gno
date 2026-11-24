# Cycle guard in the realm save walk

## Context

`saveUnsavedObjectRecursively` saves an object's unsaved children before the
object itself, and skips a child that is escaped. That skip is the only thing
that stops the walk at a reference cycle. `incRefCreatedDescendants`
escape-marks a node the created-graph walk reaches twice, so a cycle of new
objects normally holds one escaped node.

A deletion in the same transaction can leave a cycle with none, in two ways:

- `FinalizeRealmTransaction` runs `processNewDeletedMarks` before
  `processNewEscapedMarks`, and the second demotes every mark whose refcount
  reads `<= 1`. A new cycle attached under a stored object that the
  transaction then deletes drops back to refcount 1 and loses its only mark.
- A stored object deleted and then pointed at by a new object it points back
  to is revived by `incRefCreatedDescendants`, which gives it an owner and no
  escape mark, so the cycle through it is never marked.

The save walk then recurses until Go's stack limit. That ends in `fatal error:
stack overflow`, a `runtime.throw` that no `recover` reaches, so every node
executing the transaction exits. The walk charges no gas, so gas does not bound
it.

## Decision

`saveUnsavedObjects` builds a set of the objects on the save walk's current
path and threads it through `saveUnsavedObjectRecursively`. An object reached
while already on the path closes a cycle with no escaped node. The guard marks
it through `MarkNewEscaped`, clears its owner and returns:

- the descendant being saved serializes its reference through `toRefValue`'s
  NewEscaped branch, by ObjectID and with no hash;
- the object's own pending frame then saves it, and `saveObject` turns
  NewEscaped into Escaped, so its hash goes to the iavl store like any escaped
  object's.

The guard fires only where the walk would otherwise re-enter an object on its
own path, which on the base never terminates. Every transaction the base
completes takes the same path, so results, stored bytes and gas are unchanged.

## Alternatives considered

- Decide the demotion on the refcount before deletions, or redo the
  created-graph walk after them. Either changes which objects end up escaped in
  ordinary transactions, which moves stored bytes and gas, and needs every node
  to upgrade at one height. Neither covers the revived object.
- A depth limit on the walk. It turns the crash into a refused transaction, but
  a legitimate deep graph would hit the same limit, and the limit would be a
  new consensus parameter.

## Consequences

- The demotion stays, so `rlm.escaped` can miss an object the guard escaped,
  and `rlm.newEscaped` can hold it twice, once from its demoted mark. Nothing
  reads `rlm.escaped` after the save walk, since `saveNewEscaped` is empty; its
  only readers are `debugAssert` checks that run before the walk. Only
  `clearMarks` reads `rlm.newEscaped` after this point.
- A cycle the guard escapes after its outside reference is deleted stays stored
  with refcount 1, as reference counting already leaves a cycle that keeps its
  escape mark.
- `gnovm/tests/files/zrealm_cycle_unescaped.gno` builds the three cycles, a
  self-referencing struct pointer, a map containing itself and a revived stored
  object, and pins the stored objects in its `// Realm:` golden.
  `zrealm_cycle_unescaped_later.gno` reads, writes, wraps and drops a
  guard-escaped cycle in later finalizes of the same transaction, and
  `gno.land/pkg/integration/testdata/realm_cycle_guard_restart.txtar` loads a
  guard-escaped cycle and a revived one back from the store across node
  restarts, writes into both and deletes them.
