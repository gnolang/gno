# Interrealm scenario inventory (T29 working file)

Working map from every interrealm scenario in the code base to the patterns
`a`–`m` of `docs/resources/gno-interrealm-whitepaper.md`, so the next doc round
can add the scenarios the whitepaper does not show yet.

A *scenario* is one distinct situation: who calls whom (user, run script,
realm, `/p/`, stdlib), how (crossing call, plain call, method, closure, bound
method value, declared fn as value, interface, callback, `defer`,
panic/recover), what is touched (own/other realm object, pointer/value,
slice/map/struct, package var, fresh object), and the outcome (which write
lands, or the exact error start).

Every row was checked against the fixture's own golden (`// Output:` /
`// Error:`), not guessed. Function-value rows whose outcome depends on S3 or
R4 are marked **fn-rule** and are re-checked when either changes.

Legend

- Rule: R1 cross to write, R2 only pointers cross, S3 a function value writes
  only in the crossing call that made it, R4 makers stay home, R6 one import
  rule.
- Pattern: `a`–`m` as numbered in whitepaper §5; `MISSING` = no pattern shows
  this situation (see §16).
- Outcome: the printed write, or the first line of the golden error, quoted.
  `OUT:` abbreviates `// Output:`, `ERR:` abbreviates `// Error:`.
- `fn-rule` is appended to the Rule cell of every row whose outcome depends on
  S3 or R4 (a maker, an adopted value, a stored value); these rows must be
  re-checked when either rule changes.

Sources read: 460 `gnovm/tests/files/zrealm_*.gno`; 92 other filetests whose
body involves realms/crossing; 56 `examples/gno.land/r/tests/vm/**` fixtures
and 18 `examples/gno.land/p/demo/tests/**` helpers; 60 integration txtars out
of 263 (the interrealm subset); 61 non-test `.gno` realm files (36 realm
directories) under `examples/gno.land/r`.

## 1. Cross-call syntax and gates (`zrealm_cross_*`)

| Scenario | Fixtures | Outcome | Rule | Pattern |
|---|---|---|---|---|
| Crossing fn called with bare `cross` and with `cross(cur)`; both print cur path | zrealm_cross_basic | OUT: bare cross: gno.land/r/demo/crossbasic | R1 | a |
| Closure that captured `cur` is called, then crosses | zrealm_cross_closurecap | OUT: gno.land/r/demo/crossclosurecap (write lands) | R1, S3 fn-rule | a, e |
| Cross made from a sibling closure that captured another frame's cur | zrealm_cross_closuresib | ERR: cross: rlm is not the current cur (stale capture or sibling frame) | R1 | a |
| Cross into a realm of another package (crosspkg) and print cur/prev | zrealm_cross_crosspkg | OUT: current realm: CodeRealm{... gno.land/r/tests/vm/crossrealm } | R1 | a |
| `defer f(cross(rlm))` — crossing call deferred directly | zrealm_cross_defer | ERR: crossing call `cross(rlm)` cannot be deferred directly | R1 | a |
| Same, with the deferred call in another realm's function | zrealm_cross_defercross | ERR: crossing call `cross(rlm)` cannot be deferred directly | R1 | a |
| Crossing call made inside a deferred closure after the body | zrealm_cross_deferwrap | OUT: body done / from defer: gno.land/r/demo/crossdeferwrap | R1 | a, e |
| `cross(a, b)` — two arguments | zrealm_cross_extra | ERR: cross takes exactly one argument: the in-scope realm | R1 | a |
| `cross()` — no argument | zrealm_cross_noarg | ERR: cross takes exactly one argument: the in-scope realm | R1 | a |
| `cross(rlm)` on a function that has no `cur` first parameter | zrealm_cross_nonxfn | ERR: cross(rlm) can only be used as the first argument to a crossing function | R1 | a |
| `cross(rlm)` where rlm is returned by a helper (not a bare identifier) | zrealm_cross_notname | ERR: cross argument must be a bare realm-typed identifier | R1 | a |
| `cross(cur.Previous())` | zrealm_cross_previous | ERR: cross argument must be a bare realm-typed identifier | R1 | a |
| `rlm` parameter threaded into `cross(rlm)` from a helper | zrealm_cross_rlm | OUT: via helper: gno.land/r/demo/crossrlm | R1 | a |
| Stale realm (captured in an outer frame) handed to cross | zrealm_cross_stalerlm | ERR: cross: rlm is not the current cur (stale capture or sibling frame) | R1 | a |
| `cross(rlm)` used as a statement, not as first call argument | zrealm_cross_stray | ERR: cross(rlm) can only be used as the first argument to a crossing function | R1 | a |
| Cross through an interface method with `cur` first parameter | zrealm_cross_iface | OUT: iface:gno.land/r/demo/crossiface | R1, R6 | a, j |
| Cross through an interface holding a stale cur | zrealm_cross_ifacestale | ERR: cross: rlm is not the current cur (stale capture or sibling frame) | R1, R6 | a, j |
| Old `cross1` builtin removed from the language | zrealm_cross1_removed | ERR: name cross1 not declared | R1 | a |
| Panic inside a crossing call; a defer that would cross again | zrealm_crosspanic0 | ERR: fail: the transaction should have been aborted; this defer should not even be called | R1, R2 | l |

## 2. Deployed-realm interplay (`zrealm_crossrealm*`)

| Scenario | Fixtures | Outcome | Rule | Pattern |
|---|---|---|---|---|
| Foreign-realm-typed package var declared in the caller | zrealm_crossrealm0 | ERR: cannot allocate gno.land/r/tests/vm.TestRealmObject in realm gno.land/r/crossrealm | R2 | d |
| Same, initialized in init instead of a var declaration | zrealm_crossrealm1 | ERR: cannot allocate gno.land/r/tests/vm.TestRealmObject in realm gno.land/r/crossrealm | R2 | d |
| Same, allocated with new() in the caller | zrealm_crossrealm2 | ERR: cannot allocate gno.land/r/tests/vm.TestRealmObject in realm gno.land/r/crossrealm | R2 | d |
| Same, via composite literal of a foreign declared struct | zrealm_crossrealm3 | ERR: cannot allocate gno.land/r/tests/vm.TestRealmObject in realm gno.land/r/crossrealm | R2 | d |
| Write a field of a foreign-realm object held as a value through a crossing call | zrealm_crossrealm4 | OUT: modify: gno.land/r/tests/vm's code was entered without crossing and is read-only | R1, R2 | d |
| Same with a second order (4b) — object printed after the refused write | zrealm_crossrealm4b | OUT: modify: gno.land/r/tests/vm's code was entered without crossing and is read-only | R1, R2 | d |
| Write through a foreign package var of a declared type | zrealm_crossrealm5 | ERR: cannot directly modify readonly tainted object (call a crossing function of the owning realm): somevalue | R1, R2 | b, d |
| Crossing call into `p/demo/tests/v0` mutates its object and prints it | zrealm_crossrealm6 | OUT: (struct{("modified" string)} gno.land/p/demo/tests/v0.TestRealmObject) | R1 | a, c |
| Write a foreign object's field from library code | zrealm_crossrealm7 | ERR: cannot modify object owned by another realm from library code (call a function of the owning realm): somevalue | R1 | c |
| Same through an exported package var of the foreign realm | zrealm_crossrealm8 | ERR: cannot modify object owned by another realm from library code (call a function of the owning realm): SomeValue | R1 | c |
| Same through a `const`-like package reference | zrealm_crossrealm9 | ERR: cannot directly modify readonly tainted object (call a crossing function of the owning realm): (const (ref(gno.land/p/demo/tests/v0) package{})) | R1, R2 | b |
| Same for a foreign field reached through a nested value | zrealm_crossrealm10 | ERR: cannot modify object owned by another realm from library code (call a function of the owning realm): SomeValue | R1 | c |
| Print `PreviousRealm`/`IsUser`/`IsUserCall` stack table from a crossing call | zrealm_crossrealm11 | OUT: STACK = unsafe.PreviousRealm / user1.gno !> r/crossrealm.main2 -> unsafe.PreviousRealm | R1 | i |
| `unsafe` imported but not declared in the file (dot-import variant) | zrealm_crossrealm12 | ERR: name unsafe not declared | R1 | i |
| Print CurrentRealm/PreviousRealm from a plain (non-crossing) function | zrealm_crossrealm13 | OUT: From main: PR() CurrentRealm: CodeRealm{...} | R1 | b, i |
| Same, with a crossing `PrintRealm(cur realm)` helper | zrealm_crossrealm13a | OUT: From main: PrintRealm: CurrentRealm: CodeRealm{...} | R1 | a, i |
| Print cur/prev through a render helper taking `cur` and formatting addresses | zrealm_crossrealm13a_cur | OUT: PrintRealm: PreviousRealm: UserRealm{...} | R1 | a, i |
| Foreign method takes a pointer to the caller's object and writes it | zrealm_crossrealm14 | ERR: cannot modify object owned by another realm from library code (call a function of the owning realm): c<VPBlock | R1 | c, j |
| Crossing method `Foo(cur realm)` on a foreign object, then plain `Bar` | zrealm_crossrealm15 | OUT: init / hello gno.land/r/crossrealm | R1, R6 | j |
| Same shape without crossing (plain method call on foreign object) | zrealm_crossrealm15b | ERR: cannot cur-call to external realm function gno.land/r/crossrealm.Foo from gno.land/r/tests/vm | R1, R6 | j |
| Crossing method call from main, printing three cur values | zrealm_crossrealm16 | OUT: hello gno.land/r/crossrealm / hello gno.land/r/tests/vm/crossrealm | R1 | j |
| Plain (non-crossing) call of the foreign crossing method | zrealm_crossrealm16b | ERR: cannot cur-call to external realm function gno.land/r/crossrealm.Foo from gno.land/r/tests/vm | R1 | j |
| Crossing method on a `container` type holding a foreign object | zrealm_crossrealm17 | OUT: hello container gno.land/r/crossrealm | R1 | j |
| Same container shape, plain call of Foo | zrealm_crossrealm17b | ERR: cannot cur-call to external realm function gno.land/r/ | R1 | j |
| Same container shape, value receiver crossed from another type | zrealm_crossrealm17c | ERR: cannot cur-call to external realm function gno.land/r/crossrealm.Foo | R1 | j |
| Crossing method from a stored object in init, then from main | zrealm_crossrealm18 | OUT: init / hello gno.land/r/crossrealm | R1 | j |
| Non-crossing call of the same stored object's method | zrealm_crossrealm18b | ERR: cannot cur-call to external realm function gno.land/r/crossrealm.Foo | R1 | j |
| Same with the object reached through a pointer field | zrealm_crossrealm18c | ERR: cannot cur-call to external realm function gno.land/r/crossrealm.Foo | R1 | j |
| Crossing method Foo then plain Bar, recovered with `try` | zrealm_crossrealm19 | OUT: hello A gno.land/r/crossrealm / bar: gno.land/r/crossrealm's code was entered without crossing | R1 | j, l |
| Stored method value of a foreign crossing method called from the caller | zrealm_crossrealm20 | OUT: hello Foo gno.land/r/crossrealm / bar: ... was entered without crossing | R1 | f, j |
| Chain A -> B (`crossrealm_b`) printing cur and prev | zrealm_crossrealm21 | OUT: hello A cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm | R1 | a, i |
| Chain A -> B -> C printing cur/prev at each hop | zrealm_crossrealm22 | OUT: hello C cur=gno.land/r/tests/vm/... | R1 | a, i |
| Print previous address and pkgpath across a two-realm chain | zrealm_crossrealm23 | OUT: g1kscy... gno.land/r/tests/vm / g1h2y7... gno.land/r/crossrealm | R1 | i |
| Pass `&struct{A int}` pointer into another realm's gate and read it back | zrealm_crossrealm24 | OUT: &(struct{(123 int)} gno.land/r/crossrealm.Struct) 1 | R2 | d |
| Set a field of the caller's struct through the other realm's gate | zrealm_crossrealm25 | OUT: s := &Struct{A: 100} / crossrealm_b.SetObject(cross, &s.A) | R1, R2 | a, d |
| Stored pointer then incremented by the other realm's gate | zrealm_crossrealm25a | ERR: cannot directly modify readonly tainted object (call a crossing function of the owning realm): *(ptr<VPBlock | R1, R2 | b |
| Same as 25 but with a local field write between the two crossing calls | zrealm_crossrealm25b | OUT: s.A = 123 / &(struct{(123 int)} gno.land/r/crossrealm.Struct) | R1, R2 | d |
| Stored global pointer written after being handed to the other realm | zrealm_crossrealm25c | OUT: s_g = s / s_g.A = 123 | R1, R2 | b |
| Pointer to a struct field of another realm kept after the call | zrealm_crossrealm26 | OUT: &(struct{(123 int)} gno.land/r/crossrealm.Struct) | R2 | b |
| Same, object created in init and mutated later | zrealm_crossrealm27 | OUT: &(struct{(123 int)} gno.land/r/crossrealm.Struct) | R2 | b |
| Same, object passed through two realms | zrealm_crossrealm28 | OUT: &(struct{(123 int)} gno.land/r/crossrealm.Struct) | R2 | b |
| Same, pointer stored in a slice before the crossing call | zrealm_crossrealm29 | OUT: &(struct{(123 int)} gno.land/r/crossrealm.Struct) | R2 | b |
| Pass a foreign object's field pointer into a non-crossing setter | zrealm_crossrealm30 | ERR: sa.A should not be mutable; it should be attached to crossrealm_b. | R1, R2 | b |
| Same with the setter returning a struct value | zrealm_crossrealm30b | ERR: sa.A should not be mutable; it should be attached to crossrealm_b. | R1, R2 | b |
| Same with a named struct result | zrealm_crossrealm30c | ERR: sa.A should not be mutable; it should be attached to crossrealm_b. | R1, R2 | b |
| Same setter made crossing `Set(cur realm, sa *StructA)` | zrealm_crossrealm30d | OUT: Success: sa direct manipulation succeeded, meaning it was attached to the callee | R1 | a |
| Stored func value in a registry map, called with cross | zrealm_crossrealm31 | OUT: current realm: CodeRealm{...} / previous realm: CodeRealm{...} | R1 | f |
| Same registry reached through a second realm | zrealm_crossrealm31a | OUT: current realm: CodeRealm{ g1kscy... } / previous realm: CodeRealm{ g1kscy... } | R1 | f |
| Public crossing gate stores a func value handed in from main | zrealm_crossrealm32 | OUT: current realm: CodeRealm{...} / previous realm: CodeRealm{...} | R4 fn-rule | f, g |
| Same, value re-called after the storing call returned | zrealm_crossrealm32a | OUT: current realm: CodeRealm{ g1h2y7... } / previous realm: CodeRealm{ g1kscy... } | R4 fn-rule | f, g |
| Plain call of another realm's exported gate (no cross) | zrealm_crossrealm33 | ERR: gno.land/r/tests/vm/crossrealm_b's code was entered without crossing and is read-only (call its crossing function) | R1 | a |
| Run-script style main crosses into two realms, prints markers | zrealm_crossrealm34 | OUT: main() / done | R1 | a, k |
| Panic caught, state set and revived; counter read after | zrealm_crossrealm35 | OUT: counter: 100 / panic caught: gno.land/r/tests/vm/crossrealm_d's code was en | R1 | l |
| State change attempted after the panic was caught | zrealm_crossrealm36 | OUT: counter: 100 then ERR: crossrealm_d's code was entered without crossing | R1 | l |
| Attacker recovers the refusal and checks the owner did not change | zrealm_crossrealm37 | OUT: owner: g1dao_address_here / exception caught: ... / owner == attack | R1 | l |
| Crossing gate called twice from init | zrealm_crossrealm38 | OUT: 2 | R1 | a |

## 3. Cross-to-write: realm a's gate runs caller-supplied values (`zrealm_ctw_*`)

| Scenario | Fixtures | Outcome | Rule | Pattern |
|---|---|---|---|---|
| Gate `Add(cur, n)` writes its own counter and a stored object | zrealm_ctw_cross | OUT: 3 2 | R1 | a |
| Gate calls a plain callback supplied by the caller | zrealm_ctw_callback | ERR: gno.land/r/ctw_callback's code was entered without crossing and is read-only (call its crossing function) | R1 | e |
| Gate calls a crossing callback with `cross(cur)` | zrealm_ctw_callbackx | OUT: 1 | R1 | e |
| Gate runs its own stored callbacks (own table, own helpers) | zrealm_ctw_owncb | OUT: count: 2 / boxes: 6 | R1 | c, f |
| Non-crossing method of a stored object writes its field | zrealm_ctw_method | ERR: gno.land/r/tests/vm/ctw's code was entered without crossing and is read-only (call its crossing function): l<V | R1 | j |
| Caller-supplied callback stored then fired later | zrealm_ctw_pending | ERR: gno.land/r/tests/vm/ctw's code was entered without crossing and is read-only (call its crossing function): l<V | S3 fn-rule | g |
| Caller holds a bound setter from the victim and calls it | zrealm_ctw_boundsetter | ERR: gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function) | R4 fn-rule | f |
| Value conversion of a library callback type passed through the gate | zrealm_ctw_conv | OUT: 3 7 | R1 | c |
| Gate iterates a caller-supplied container of callbacks | zrealm_ctw_iter | ERR: gno.land/r/tests/vm/ctw's code was entered without crossing and is read-only (call its crossing function): it< | R1 | c, e |
| Callback handed to the gate by the caller, fired inside the gate | zrealm_ctw_handedcb | ERR: gno.land/r/tests/vm/ctw's code was entered without crossing and is read-only (call its crossing function): (co | R1 | e |
| Gate constructor runs a callback before the object is stored | zrealm_ctw_ctor | OUT: 5 then ERR: ... entered without crossing: l<V | R1 | e |
| Gate recovers a panic raised by library code and prints the value | zrealm_ctw_recover | OUT: recovered: boom / count: 1 / value: boom | R1 | l |
| Render path calls the gate and prints rows | zrealm_ctw_render | OUT: rows:xx | R1 | a, k |

## 4. `cur` identity, binding and threading (`zrealm_cur_*`)

| Scenario | Fixtures | Outcome | Rule | Pattern |
|---|---|---|---|---|
| Print cur address and pkgpath, compare identity | zrealm_cur_smoke | OUT: cur == cur_: true / cur.PkgPath: gno.land/r/curtest | R1 | i |
| A -> B -> A chain: cur/prev pkgpath at each hop | zrealm_cur_chain | OUT: A prev.PkgPath: gno.land/r/curchaintest | R1 | i |
| Classify cur vs unsafe runtime: IsCode/IsUser/IsUserCall/IsUserRun | zrealm_cur_classify | OUT: cur.IsUserCall == rt.IsUserCall: true | R1 | i |
| Compare cur-derived values with runtime-derived ones after crossing | zrealm_cur_parity | OUT: init parity: true / main parity: true | R1 | i |
| `cur.Previous()` for an EOA (user call) | zrealm_cur_eoa | OUT: pkgPath match (both empty): true | R1 | i |
| `func init(cur realm)` sees itself as cur, EOA as previous | zrealm_cur_init | OUT: init cur.PkgPath: gno.land/r/curinit | R1 | MISSING |
| `main(cur realm)` in a run script, fields and previous | zrealm_cur_classify, zrealm_cur_eoa | OUT: cur.IsUserRun == rt.IsUserRun: true | R1 | k |
| Thread `cur` into a same-package helper (legal) | zrealm_cur_legal | OUT: threaded: gno.land/r/demo/curlegal | R1 | a |
| Alias a realm type `type R = realm` and pass `cur` as `R` | zrealm_cur_alias_legal | OUT: aliased: gno.land/r/demo/curaliasok | R1 | MISSING |
| `cur` in a `for ... range` body (legal read) | zrealm_cur_legal | OUT: range-define pkgpath: / int cur: 7 | R1 | i |
| Deferred closure that reads cur after the body | zrealm_cur_defer_legal | OUT: defer IsCurrent: true / defer PkgPath: gno.land/r/demo/curdeferok | R1 | i |
| Crossing method declared on a type, applied to a value | zrealm_cur_method | OUT: non-cross: t.M(cur) returns same realm as caller cur: true | R1 | j |
| Method expression with a crossing method | zrealm_cur_method_expr | OUT: m: real gno.land/r/demo/curmethexpr true | R1 | j |
| Crossing method called through an interface value | zrealm_cur_method_legal | OUT: inner IsCurrent: true | R1 | j, h |
| Address of `cur` taken | zrealm_cur_addr | ERR: cannot take the address of a realm-typed `cur` | R1 | i |
| Address of `cur` taken through a pointer argument helper | zrealm_cur_addr_arg | ERR: cannot take the address of a realm-typed `cur` | R1 | i |
| Realm type alias as first parameter name (`func f(c R, cur R)`) | zrealm_cur_alias | ERR: a crossing function's first realm argument must have name cur | R1 | MISSING |
| Second realm-typed parameter named cur | zrealm_cur_alias_second | ERR: only the first realm type argument of a crossing function | R1 | MISSING |
| Reassign `cur` inside the crossing function | zrealm_cur_reassign | ERR: cannot reassign the crossing `cur` parameter | R1 | i |
| Shadow `cur` in an inner block | zrealm_cur_shadow | ERR: cannot reassign the crossing `cur` parameter | R1 | i |
| Pass a shadowed `cur` to a cross-call | zrealm_cur_shadow_call | ERR: only the `cur` argument of a containing crossing function maybe passed by cross-call | R1 | i |
| Same inside an if-block | zrealm_cur_shadow_if | ERR: only the `cur` argument of a containing crossing function maybe passed by cross-call | R1 | i |
| Same inside a range body | zrealm_cur_shadow_range | ERR: only the `cur` argument of a containing crossing function maybe passed by cross-call | R1 | i |
| Assign to `cur` in a range clause | zrealm_cur_range | ERR: cannot assign to a realm-typed `cur` in a range clause | R1 | i |
| Persist `cur` in a package var | zrealm_cur_persist_var | ERR: cannot persist realm value: realm values are ephemeral and tied to a call frame | R1 | MISSING |
| Persist `cur` inside a stored struct | zrealm_cur_persist_struct | ERR: cannot persist realm value: realm values are ephemeral and tied to a call frame | R1 | MISSING |
| Persist `cur` inside a closure stored in state | zrealm_cur_persist_closure | ERR: cannot persist realm value: realm values are ephemeral and tied to a call frame | R1 | MISSING |
| Crossing function entered by a plain call (backstop) | zrealm_cur_backstop | ERR: crossing function gno.land/r/demo/curbackstop.target was entered without cross(), so it takes its caller's identity | R1 | a |
| Same, entered from a deferred plain call | zrealm_cur_defer | ERR: crossing function gno.land/r/demo/curdefer.target was entered without cross(), so it takes its caller's identi | R1 | a |
| Same, as a method | zrealm_cur_method_backstop | ERR: crossing function gno.land/r/demo/curmethbs.target was entered without cross(), so it takes its caller's ident | R1 | a, j |

## 5. Entry points and library code touching a realm object (`zrealm_entry_*`)

| Scenario | Fixtures | Outcome | Rule | Pattern |
|---|---|---|---|---|
| Caller writes a field of the victim's object passed as an arg | zrealm_entry_arg_write | OUT: copy to victim: cannot modify object owned by another realm from library code | R1 | c |
| Victim's own deferred write vs foreign deferred write | zrealm_entry_defer | OUT: owner deferred: v1 / foreign deferred: cannot modify object owned by another realm | R1 | c, l |
| Escaped pointer from the victim written by caller code | zrealm_entry_escaped_ptr | OUT: escaped: cannot modify object owned by another realm from library code | R1 | b, c |
| Fresh foreign object created by the caller and written | zrealm_entry_fresh_object | OUT: fresh foreign: cannot modify object owned by another realm from library code | R1 | c |
| Interface value holding the victim's object, method called | zrealm_entry_iface | OUT: owner wrote: v1 then ERR: cannot modify object owned by another realm | R1, R6 | h |
| Type assertion to a concrete victim type then field write | zrealm_entry_iface_assert | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1 | h |
| Asserted type written inside a deferred closure | zrealm_entry_iface_defer | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1 | h, l |
| Interface embedding another interface, method resolution | zrealm_entry_iface_embed | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1, R6 | h |
| Embedded pointer type satisfying the interface | zrealm_entry_iface_embedptr | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1, R6 | h |
| Interface produced by a generic constructor | zrealm_entry_iface_gen | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1, R6 | h |
| Interface value stored in a map, then called | zrealm_entry_iface_map | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1, R6 | h |
| Interface made by a `func() I` helper | zrealm_entry_iface_mkh | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1, R6 | h |
| Method value taken from an interface value and called | zrealm_entry_iface_mv | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1, R6 | h |
| Method value from an interface applied by a library `Apply` | zrealm_entry_iface_mvapply | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R4 fn-rule | f, h |
| Method value from an interface passed as an argument | zrealm_entry_iface_mvarg | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R4 fn-rule | f, h |
| Pointer-typed interface value method value | zrealm_entry_iface_mvp | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1, R6 | h |
| Interface method value stored then called | zrealm_entry_iface_mvstored | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | S3 fn-rule | g, h |
| Named interface with a value receiver | zrealm_entry_iface_named | OUT: named: pwnd-via-iface / iface: pwnd-via-iface | R1 | h |
| Primitive-based type with a method satisfying the interface | zrealm_entry_iface_prim | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1, R6 | h |
| Pointer receiver type satisfying the interface | zrealm_entry_iface_ptr | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1, R6 | h |
| Interface value stored in the victim's state, then called | zrealm_entry_iface_stored | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1, R6 | h |
| Stored pointer-to-interface | zrealm_entry_iface_storedptr | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1, R6 | h |
| Stored interface holding a value (not pointer) | zrealm_entry_iface_storedval | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1, R6 | h |
| Sub-interface of the victim's interface | zrealm_entry_iface_sub | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1, R6 | h |
| Type switch over several interface implementations | zrealm_entry_iface_switch | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,1)>.Field | R1, R6 | h |
| Interface value on an object that was never stamped by its realm | zrealm_entry_iface_unstamped | ERR: cannot modify object owned by another realm from library code: tree<VPBlock(1,0)>.node | R1, R6 | h |
| Multi-method interface, one method writes | zrealm_entry_iface_multi | OUT: hi / victim: victim-orig (write refused, state intact) | R1, R6 | h |
| Interface wrapper delegating directly to the victim object | zrealm_entry_iface_wrap_direct | OUT: direct | R1 | h |
| Interface wrapper holding an interface value | zrealm_entry_iface_wrap_iface | OUT: iface | R1 | h |
| Interface wrapper holding an interface value with a value receiver | zrealm_entry_iface_wrap_ifacev | OUT: ifacev | R1 | h |
| Bound method value of the victim's object called by a stranger | zrealm_entry_method_value | OUT: stranger call: cannot modify object owned by another realm from library code | R4 fn-rule | f |
| `mutable()` builtin no longer exists | zrealm_entry_no_mutable | ERR: name mutable not declared | R1 | MISSING |
| Library `call(cur, f)` runs a caller closure that writes the caller's realm | zrealm_entry_p_closure | OUT: p closure: cannot modify object owned by another realm from library code: b<~VPBlock(1,0)>.N | R1 | e |
| Nested closure literal in library code writing a captured victim object | zrealm_entry_p_closure_nested | OUT: lit: gno.land/r/demo/entrynestclosure's code was entered without crossing | R1 | e |
| Library handle taking the victim object and calling its writer | zrealm_entry_p_handle | OUT: writer: cannot modify object owned by another realm from library code | R1 | c, h |
| Library declares an anonymous writer fn over the victim's type | zrealm_entry_pfunc_anon | ERR: cannot modify object owned by another realm from library code: i<VPBlock(1,0)>.Field | R1 | f |
| Library writer takes `[]any` and writes through it | zrealm_entry_pfunc_anyslice | ERR: cannot modify object owned by another realm from library code: i<VPBlock(1,0)>.Field | R1 | f |
| Library writer over an array of the victim's type | zrealm_entry_pfunc_arr | ERR: cannot modify object owned by another realm from library code: i<VPBlock(1,0)>.Field | R1 | f |
| Victim's declared writer handed to `pfuncbind.Apply` | zrealm_entry_pfunc_echo | ERR: cannot modify object owned by another realm from library code: i<VPBlock(1,0)>.Field | R4 fn-rule | f |
| Same through a map of function values | zrealm_entry_pfunc_map | ERR: cannot modify object owned by another realm from library code: i<VPBlock(1,0)>.Field | R4 fn-rule | f |
| Named function value stored in the victim's package var | zrealm_entry_pfunc_pkgvar | ERR: cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,0)>.Field | S3 fn-rule | g |
| Receiver method value handed to a library `Apply` | zrealm_entry_pfunc_recv | ERR: cannot modify object owned by another realm from library code: i<VPBlock(1,0)>.Field | R4 fn-rule | f |
| Receiver method value with a receiver-typed argument | zrealm_entry_pfunc_recvr | ERR: cannot modify object owned by another realm from library code: i<VPBlock(1,0)>.Field | R4 fn-rule | f |
| Same-package (p) writer applied to the victim's object | zrealm_entry_pfunc_samepkg | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,0)>.Field | R1, R6 | f, h |
| Writer over a slice of the victim's type | zrealm_entry_pfunc_slice | ERR: cannot modify object owned by another realm from library code: i<VPBlock(1,0)>.Field | R1 | f |
| Victim sorts its own slice of its own objects | zrealm_entry_pfunc_sortown | OUT: slice[(1 int),(2 int),(3 int)] (write lands) | R1 | c |
| Named function value applied from a package var | zrealm_entry_pfunc_var | ERR: cannot modify object owned by another realm from library code: i<VPBlock(1,0)>.Field | R4 fn-rule | f |
| Library writer obtained from a view-only caller | zrealm_entry_pfunc_viewonly_from | ERR: cannot modify object owned by another realm from library code: i<VPBlock(1,0)>.Field | R1, R6 | f, h |
| Promoted method through an embedded victim object | zrealm_entry_promoted | OUT: owner wrote: v1 / promoted: cannot modify object owned by another realm | R1 | j |
| Victim's registry of crossing functions called with `cross(cur)` | zrealm_entry_registry_fn | OUT: start: 0 / after: 5 | R1 | f |

## 6. Function values and declared values crossing (`zrealm_fv_*`)

| Scenario | Fixtures | Outcome | Rule | Pattern |
|---|---|---|---|---|
| Function value assigned to a new variable, then called | zrealm_fv_var | ERR: cannot allocate gno.land/r/tests/vm/fvx.T in realm gno.land/r/fvvar | R2 | d |
| Function value call result assigned to an alias type | zrealm_fv_alias_anon | OUT: 0 | R2 | d |
| Alias type variable from an anonymous-typed literal | zrealm_fv_alias_anon_lit | OUT: 0 1 2 0 | R2 | d |
| Function-value literal passed as an argument across the call | zrealm_fv_anon_arg | OUT: 5 | S3 fn-rule | f |
| Same, non-crossing variant | zrealm_fv_anon_arg_nc | OUT: 5 | S3 fn-rule | f |
| Anonymous struct field of the victim's declared type | zrealm_fv_anon_field | ERR: cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d |
| Anonymous literal created by the owning realm and written | zrealm_fv_anon_own | OUT: 9 / 7 | R1 | d |
| Anonymous literal returned by a function | zrealm_fv_anon_ret | OUT: 3 / 3 | R2 | d |
| Anonymous struct with an unexported field copied out | zrealm_fv_anon_unexported | ERR: cannot copy struct{n int} outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d |
| `any` holding a pointer to the victim's value | zrealm_fv_anyp | OUT: 1 | R2 | d |
| `any` holding an interface value | zrealm_fv_anyp_iface | OUT: 1 | R2 | h |
| `any` with an untyped nil | zrealm_fv_anyp_nil | OUT: true | R2 | d |
| `any` holding a value of the victim's declared type | zrealm_fv_anyp_val | ERR: cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d |
| Write through `any` holding the victim's pointer | zrealm_fv_anyp_write | ERR: cannot directly modify readonly tainted object (call a crossing function of the owning realm): x<VPBlock(1,1)> | R1, R2 | b |
| `append(spread...)` of the victim's values | zrealm_fv_append_spread | ERR: cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d |
| Array of the victim's named type allocated by the caller | zrealm_fv_arr | ERR: cannot allocate gno.land/r/tests/vm/fvx.H in realm gno.land/r/fvarr | R2 | d |
| Array of an anonymous struct of the victim's type | zrealm_fv_arr_anon | ERR: cannot allocate gno.land/r/tests/vm/fvx.T in realm gno.land/r/fvarranon | R2 | d |
| Victim value boxed inside `any` then copied | zrealm_fv_box | ERR: cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d |
| Declared function value called after copy | zrealm_fv_call | OUT: 1 1 | S3 fn-rule | f |
| Conversion to the victim's declared type | zrealm_fv_conv | ERR: illegal conversion to external realm type | R6 | d, h |
| Copy of the victim's value into `any` | zrealm_fv_copy_any | ERR: cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d |
| Copy through `append` | zrealm_fv_copy_append | ERR: cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d |
| Copy by assignment | zrealm_fv_copy_assign | ERR: cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d |
| Copy out, change, hand `&c` back to the owner's crossing function | zrealm_attack_copy_handback | ERR: cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d |
| Method value copies its receiver | zrealm_fv_copy_mv | ERR: cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d |
| Method value through an interface copies the receiver | zrealm_fv_copy_mv_iface | ERR: cannot copy gno.land/r/tests/vm/fvx.E outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d, h |
| Copy from a view-only (readonly) object | zrealm_fv_copy_viewonly | ERR: cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d |
| Function value cost measurement (gas) | zrealm_fv_cost | OUT: 1 | S3 fn-rule | f |
| Function value in a defer argument | zrealm_fv_defer_val | ERR: cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d, l |
| Error value of the victim's declared error type defined | zrealm_fv_err_define | ERR: cannot copy gno.land/r/tests/vm/fvx.E outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d |
| Local conversion into the victim's type | zrealm_fv_forge_local | ERR: illegal conversion to external realm type | R6 | d |
| Own type converted locally and written (legal) | zrealm_fv_forge_own | OUT: 7 | R1 | d |
| Conversion from a readonly value | zrealm_fv_forge_ro | ERR: illegal conversion to external realm type | R6 | d |
| Iterate a `map[K]*T` of the victim's type | zrealm_fv_iter_ptr | OUT: 1 | R2 | d |
| Iterate a `map[K]T` (value entries) | zrealm_fv_iter_value | ERR: cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d |
| `make([]T, ...)` of the victim's type | zrealm_fv_make | ERR: cannot allocate gno.land/r/tests/vm/fvx.T in realm gno.land/r/fvmake | R2 | d |
| Method expression of the victim's method | zrealm_fv_methexpr | OUT: 1 | S3 fn-rule | f |
| `map[k]func()` lookup that misses | zrealm_fv_miss | OUT: 0 false | S3 fn-rule | f |
| `new(T)` of the victim's type | zrealm_fv_new_anon | ERR: cannot allocate gno.land/r/tests/vm/fvx.T in realm gno.land/r/fvnewanon | R2 | d |
| Own value and own function value mixed | zrealm_fv_own | OUT: 3 1 200 | R1 | f |
| Panic then function-value calls in recover | zrealm_fv_panic_then_call | OUT: -1 / 1 2 / deferred closure 1 | S3 fn-rule | f, l |
| Parenthesized method value | zrealm_fv_paren_mv | OUT: 1 | S3 fn-rule | f |
| Range with blank identifier over a func slice | zrealm_fv_range_blank | OUT: 1 | S3 fn-rule | f |
| Range with index over a func slice | zrealm_fv_range_idx | OUT: 303 | S3 fn-rule | f |
| Function value copied inside a recovering defer | zrealm_fv_recover | ERR: cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d, l |
| Receiver method value from an interface | zrealm_fv_recv_iface | OUT: x | S3 fn-rule | f, h |
| Argument evaluation order when a copy is involved | zrealm_fv_recv_order | OUT: 100 / 100 | R2 | d |
| Named anonymous-struct result with a pointer to its field | zrealm_fv_res_anon | ERR: cannot allocate gno.land/r/tests/vm/fvx.T in realm gno.land/r/fvresanon | R2 | d |
| Array result with a pointer into it | zrealm_fv_res_arr | ERR: cannot allocate gno.land/r/tests/vm/fvx.T in realm gno.land/r/fvresarr | R2 | d |
| Named type result with a pointer to it | zrealm_fv_res_h | ERR: cannot allocate gno.land/r/tests/vm/fvx.H in realm gno.land/r/fvresh | R2 | d |
| Slice capacity growth on a foreign slice | zrealm_fv_slice_cap | ERR: cannot allocate gno.land/r/tests/vm/fvx.T in realm gno.land/r/fvslicecap | R2 | d |
| Composite literal of the victim's type | zrealm_fv_slicelit | ERR: cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d |
| Type switch over the victim's error type | zrealm_fv_typeswitch | ERR: cannot copy gno.land/r/tests/vm/fvx.E outside realm gno.land/r/tests/vm/fvx (use a pointer) | R2 | d |
| Zero-length slice of the victim's type | zrealm_fv_zero_len | OUT: 0 | R2 | d |

## 7. Maker rules: function values and who made them (`zrealm_maker_*`)

| Scenario | Fixtures | Outcome | Rule | Pattern |
|---|---|---|---|---|
| Victim declares its own writer `set(*Immutable)`; caller applies it | zrealm_maker_ownfunc | OUT: field: own (write lands) | S3 fn-rule | f |
| Two writers adopted from another realm's `Writers()` getter | zrealm_maker_adopted_foreign | OUT: x refused: gno.land/r/tests/vm/cbvictim's code was entered without crossing | R4 fn-rule | f |
| Getter returns a declared writer of the victim | zrealm_maker_getdecl | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | f |
| Getter returns a method value of a victim object | zrealm_maker_getmv | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: t<VPBlock(1,0)>.Name | S3 fn-rule | f |
| Getter returns a function value stored in victim state | zrealm_maker_getst | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: t<VPBlock(1,0)>.Name | S3 fn-rule | f, g |
| Getter returns a value from the victim's package var | zrealm_maker_getter | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: (const (ref(...cbvictim) package{})) | S3 fn-rule | f |
| Victim hands a writer back to the caller which then calls it | zrealm_maker_handback | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | R4 fn-rule | f |
| Method value taken from an interface held by the victim | zrealm_maker_ifacemv | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: t<VPBlock(1,0)>.Name | S3 fn-rule | f, h |
| Same, with the call made through the victim's gate | zrealm_maker_ifacemv_gate | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: t<VPBlock(1,0)>.Name | S3 fn-rule | f, h |
| Method expression `T.M` of a victim type | zrealm_maker_methodexpr | ERR: cbvictim's code was entered without crossing and is read-only: t<VPBlock(1,0)>.Name | S3 fn-rule | f |
| Bound method value `obj.M` of a victim object | zrealm_maker_methodvalue | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: t<VPBlock(1,0)>.Name | S3 fn-rule | f |
| Victim's own closure (declared in victim source) called by caller | zrealm_maker_ownclosure | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: (const (ref(...cbvictim) package{})) | S3 fn-rule | f |
| Method value of a value-receiver method on a struct field | zrealm_maker_valmv | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: v<VPBlock(1,0)>.P.Name | S3 fn-rule | f |
| Same through a gate, after the struct was passed in | zrealm_maker_valmv_gate | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: v<VPBlock(1,0)>.P.Name | S3 fn-rule | f |
| Function value wrapped in another function value | zrealm_maker_wrapped | ERR: cbvictim's code was entered without crossing and is read-only: Owner<~VPBlock(3,0)> | S3 fn-rule | f |

## 8. Laundering a foreign object through methods and callbacks (`zrealm_launder_*`)

| Scenario | Fixtures | Outcome | Rule | Pattern |
|---|---|---|---|---|
| Local type with a method copying a victim object | zrealm_launder_a_value_copy | ERR: cannot copy gno.land/r/laundera.ExploiterA outside realm gno.land/r/laundera (use a pointer) | R2, R6 | h |
| Local method with a pointer parameter writing the victim object | zrealm_launder_b_pointer_write | ERR: cannot copy gno.land/r/launderb.ExploiterB outside realm gno.land/r/launderb (use a pointer) | R2, R6 | h |
| Library (p) method receiving the victim object writes a field | zrealm_launder_c_pmethod | OUT: before: original-ptr then ERR: cannot modify object owned by another realm | R1 | c, h |
| Direct field write from library code on the victim pointer | zrealm_launder_d_pattack_direct | ERR: cannot modify object owned by another realm from library code: o<VPBlock(1,0)>.Field | R1 | c |
| Write via a method on a library type holding the pointer | zrealm_launder_e_pattack_via_method | OUT: before: original-ptr then ERR: cannot modify object owned by another realm | R1 | c, h |
| Write via a method with a receiver holding the pointer | zrealm_launder_f_pattack_receiver | ERR: cannot modify object owned by another realm from library code: o<VPBlock(1,1)>.Field | R1 | c |
| Type pun over a readonly value | zrealm_launder_g_typepun | ERR: illegal conversion of readonly or externally stored value | R6 | d |
| Library callback receiving the victim pointer writes it | zrealm_launder_h_pcallback | OUT: before: original-imm then ERR: cannot modify object owned by another realm | R1 | c, e |
| Same with a nil receiver trick | zrealm_launder_i_nilrecv | OUT: before: original-imm then ERR: cannot modify object owned by another realm | R1 | c |
| Evil func, slice and map values applied to the victim object | zrealm_launder_j_otherniltypes | OUT: EvilFunc: cannot modify object owned by another realm | R1 | e, f |
| Slice parameter aliasing the victim's slice | zrealm_launder_k_sliceparam | OUT: before: original-imm then ERR: cannot modify object: s<VPBlock(1,1)>[0] | R1, R2 | c |
| `any` parameter holding the victim pointer | zrealm_launder_l_anyparam | OUT: before: original-imm then ERR: cannot modify object: t<VPBlock(1,2)>.Field | R1 | c |
| Receiver type stamped by the receiver's package | zrealm_launder_m_pstamp_recv | ERR: cannot modify object owned by another realm from library code: i<VPBlock(1,1)>.Field | R1 | c |
| Stdlib receiver (`copy` into a foreign slice) | zrealm_launder_n_stdlib_recv | ERR: cannot directly modify readonly tainted object: dst<VPBlock(1,1)>) | R1, R2 | d |
| Closure over a receiver-stamped object | zrealm_launder_o_pstamp_closure | ERR: cannot modify object owned by another realm from library code: o<VPBlock(1,0)>.Field | R1 | c |
| Pointer launder through a package var (`pdata_ptr_deref`) | zrealm_launder_pdata_ptr_deref | OUT: pwnd-via-pdata-ptr-launder (write lands) | S3 fn-rule | MISSING |
| Value launder through a package var (`pdata_value_copy`) | zrealm_launder_pdata_value_copy | OUT: pwnd-via-pdata-value-launder (write lands) | R2, R4 fn-rule | MISSING |

## 9. Laundering against rdata: stored, converted, trampolined (`zrealm_launder_rdata_*`)

| Scenario | Fixtures | Outcome | Rule | Pattern |
|---|---|---|---|---|
| Bare embedded field reached through an older stored value | zrealm_launder_rdata_bare_embed | ERR: cannot directly modify readonly tainted object: w<VPBlock(1,1)>.Bare.Field | R1, R2 | b, d |
| Bare pointer embedded in the victim object | zrealm_launder_rdata_bare_ptr | ERR: cannot directly modify readonly tainted object: w<VPBlock(1,1)>.Inner.Field | R1, R2 | b |
| Same with a value embed | zrealm_launder_rdata_bare_val | ERR: cannot directly modify readonly tainted object: w<VPBlock(1,1)>.Inner.Field | R1, R2 | b |
| Callback over a stored victim object | zrealm_launder_rdata_callback | ERR: gno.land/r/launderrh3's code was entered without crossing and is read-only: i<VPBlock(1,0)>.Field | R1 | e |
| Composite writes: field write, bare replace, pointer-slice write | zrealm_launder_rdata_composites | OUT: 1 slice[i].Field write blocked: true / 2 slice[i] = bare blocked: true | R2 | d |
| Construct a victim-stamped object outside its realm | zrealm_launder_rdata_construct | ERR: cannot allocate gno.land/r/tests/vm/launderrvictim.Immutable in realm gno.land/r/launderrh4 | R2, R6 | d |
| Conversion via chained anonymous types | zrealm_launder_rdata_conv_anon_chain | ERR: illegal conversion of readonly or externally stored value | R6 | d |
| Conversion via an anonymous struct | zrealm_launder_rdata_conv_anon_struct | ERR: illegal conversion of readonly or externally stored value | R6 | d |
| Conversion via a bare pun (same layout) | zrealm_launder_rdata_conv_bare_pun | ERR: illegal conversion of readonly or externally stored value | R6 | d |
| Conversion via an embedded pun | zrealm_launder_rdata_conv_embed_pun | ERR: illegal conversion of readonly or externally stored value | R6 | d |
| Conversion via an external alias pointer type | zrealm_launder_rdata_conv_external_alias_ptr | ERR: illegal conversion to external realm type | R6 | d |
| Named external pointer type conversion that is legal (nil checks only) | zrealm_launder_rdata_conv_external_named_ptr_valid | OUT: true / true / true | R6 | d |
| Named external struct pointer conversion | zrealm_launder_rdata_conv_external_named_struct_ptr | ERR: illegal conversion to external realm type | R6 | d |
| Plain pointer conversion to the external type | zrealm_launder_rdata_conv_external_ptr | ERR: illegal conversion to external realm type | R6 | d |
| Function-pointer conversion carrying a foreign object | zrealm_launder_rdata_conv_funcptr | ERR: gno.land/r/launderrh_conv_funcptr's code was entered without crossing | R1 | e |
| `any` boxed pointer asserted back and written | zrealm_launder_rdata_conv_iface_box | OUT: 1 var any = ptr; .(*Bare).F= blocked: true | R2 | d, h |
| Explicit interface assertion onto the victim pointer | zrealm_launder_rdata_conv_iface_explicit | ERR: cannot directly modify readonly tainted object: box<VPBlock(1,2)>.(*(typeval{...})) | R1 | b, h |
| Pointer chain conversion | zrealm_launder_rdata_conv_ptr_chain | ERR: illegal conversion of readonly or externally stored value | R6 | d |
| Slice conversion with the same underlying type | zrealm_launder_rdata_conv_slice_same_underlying | ERR: ... cannot convert s<VPBlock(1,1)> (of type []gno.land/p/demo/tests/lau | R6 | d |
| Subslice conversion | zrealm_launder_rdata_conv_subslice | ERR: illegal conversion of readonly or externally stored value | R6 | d |
| Value pun converted in place (local copy written, victim intact) | zrealm_launder_rdata_conv_val_pun | OUT: local tv.Field: local-pwnd / victim slice[0]: slice0 | R2 | d |
| Value copy then address taken | zrealm_launder_rdata_conv_val_then_addr | OUT: 3 &s[0] write blocked: true | R2 | d |
| Embedded p-field receiver write | zrealm_launder_rdata_embed_p | OUT: before: embed-orig then ERR: cannot modify object: i<VPBlock(1,0)>.Field | R1 | c |
| Exotic write shapes: `+=`, multi-return swap, `(*p).Field` | zrealm_launder_rdata_exotic | OUT: 1 += on field blocked: true / 2 multi-return swap blocked: true | R1 | b |
| `any` assertion write, function path, `[]any` element | zrealm_launder_rdata_iface | OUT: 1 any.(*Bare).Field write blocked: true | R1 | b, h |
| `any` parameter typed as the victim's pointer | zrealm_launder_rdata_l_anyparam | ERR: *gno.land/r/tests/vm/launderrvictim.Immutable is not of type *gno.land/p/demo/tests/launderpkg/v0.Immutable | R6 | h |
| `any` parameter through a readonly receiver | zrealm_launder_rdata_lr_anyparam | ERR: gno.land/r/launderrh2's code was entered without crossing: t<VPBlock(1,2)>.Field | R1 | e |
| Method value applied via a trampoline function | zrealm_launder_rdata_mv_apply_trampoline | ERR: gno.land/r/launderrh_mv_apply_trampoline's code was entered without crossing: i<VPBlock(1,0)>.Field | R1 | f |
| Attacker's stored method value fired later | zrealm_launder_rdata_mv_attacker_stored_mv | OUT: before: embed-orig then ERR: cannot modify object: i<VPBlock(1,0)>.Field | S3 fn-rule | g |
| Higher-order chain of method values | zrealm_launder_rdata_mv_higher_order_chain | OUT: before: embed-orig then ERR: cannot modify object | R1 | f |
| Immutable hook applied to a method value | zrealm_launder_rdata_mv_immhook_apply | OUT: before: embed-orig then ERR: cannot modify object | R1 | f |
| Method expression on the victim's type | zrealm_launder_rdata_mv_method_expr | ERR: cannot modify object owned by another realm from library code: i<VPBlock(1,0)>.Field | R1 | f |
| Method value on the victim's object | zrealm_launder_rdata_mv_method_value | OUT: before: embed-orig then ERR: cannot modify object | R1 | f |
| Plain hook applied to a value | zrealm_launder_rdata_mv_plainhook_apply | OUT: before: embed-orig then ERR: cannot modify object | R1 | f |
| Closure returned by the victim then called plainly | zrealm_launder_rdata_mv_returned_closure | OUT: plain closure: victim's code was entered without crossing and is read-only | R1 | e |
| Bound method value stored in the victim's registry | zrealm_launder_rdata_mv_stored_bound_mv | ERR: cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,0)>.Field | S3 fn-rule | g |
| Stored hook function value in the handler realm | zrealm_launder_rdata_mv_stored_hook | ERR: gno.land/r/launderrh_mv_stored_hook's code was entered without crossing: i<VPBlock(1,0)>.Field | R1 | g |
| Stored interface value holding a method | zrealm_launder_rdata_mv_stored_iface | OUT: before: embed-orig then ERR: cannot modify object | S3 fn-rule | g, h |
| Tampered punned method value | zrealm_launder_rdata_mv_tamper_punned_mv | ERR: illegal conversion of readonly or externally stored value | R6 | d |
| Attacker's defer writes a captured victim pointer | zrealm_launder_rdata_pdr_attacker_defer | OUT: B defer write via captured ptr blocked: true | R1 | l |
| Deferred apply of a callback over the victim object | zrealm_launder_rdata_pdr_defer_apply | ERR: gno.land/r/launderrh_pdr_defer_apply's code was entered without crossing: i<VPBlock(1,0)>.Field | R1 | l |
| Deferred callback write | zrealm_launder_rdata_pdr_defer_cb | ERR: gno.land/r/launderrh_pdr_defer_cb's code was entered without crossing | R1 | l |
| Multi-level pointer laundering | zrealm_launder_rdata_pdr_multilevel | OUT: before: ptr-orig then ERR: cannot modify object | R1 | c |
| Deferred apply of a p-function | zrealm_launder_rdata_pdr_p_defer_apply | OUT: before: ptr-orig then ERR: cannot modify object | R1 | l |
| Panic in the victim, recovered by the attacker, write attempted | zrealm_launder_rdata_pdr_p_panic | OUT: recovered: cannot modify object... / before: ptr-orig | R1 | l |
| Recover inside the victim's code, then write | zrealm_launder_rdata_pdr_p_recover | OUT: rec: cannot directly modify readonly tainted object | R1 | l |
| Recover a panic raised by the victim, no write | zrealm_launder_rdata_pdr_p_recover_pfn | OUT: rec: non-nil / after: ptr-orig | R1 | l |
| Victim-induced panic in a deferred write | zrealm_launder_rdata_pdr_panic_defer | ERR: victim-induced panic | R1 | l |
| Recover and retry the write | zrealm_launder_rdata_pdr_recover_retry | ERR: gno.land/r/launderrh_pdr_recover_retry's code was entered without crossing | R1 | l |
| Panic, recover, repanic, outer recover | zrealm_launder_rdata_pdr_repanic | OUT: 2 panic->recover->repanic->outer-recover blocked: false | R1 | l |
| Victim recovers its own exception and retries | zrealm_launder_rdata_pdr_victim_recover | OUT: retry blocked: true | R1 | l |
| Pointer field of the victim object (`*T` inside) | zrealm_launder_rdata_ptrfield_p | OUT: before: ptr-orig then ERR: cannot modify object: i<VPBlock(1,0)>.Field | R1 | c |
| Local immutable type with a setter | zrealm_launder_rdata_typepun | ERR: illegal conversion of readonly or externally stored value | R6 | d |
| Value field of the victim object | zrealm_launder_rdata_valfield_p | OUT: before: val-orig then ERR: cannot modify object | R1 | c |

## 10. Home-realm handlers, handles and hiv (`zrealm_home_*`, `zrealm_handle_*`, `zrealm_hiv_*`)

| Scenario | Fixtures | Outcome | Rule | Pattern |
|---|---|---|---|---|
| Own-realm helper called from a gate writes (legal) | zrealm_home_own_helper | OUT: owner: own-helper | R1 | a, c |
| Own table of function values fired by the gate | zrealm_home_own_table | OUT: owner: own-table | R1 | a, f |
| Known package var of the victim reached from the caller | zrealm_home_exportvar | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | a |
| Getter returns the victim's object | zrealm_home_getter | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | b |
| Gate returns an object the caller then uses | zrealm_home_gate_result | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | a |
| Victim object passed as an argument to caller code | zrealm_home_provider | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | c |
| Victim object received as a receiver | zrealm_home_recv | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | j |
| Victim object nested in a struct field | zrealm_home_nested_struct | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | c |
| Victim object inside an array argument | zrealm_home_array_arg | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | c |
| Victim object inside a slice literal | zrealm_home_slice_lit | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | c |
| Victim object in an interface argument | zrealm_home_iface_arg | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | h |
| Method value of the victim's object | zrealm_home_methodvalue | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: t<VPBlock(1,0)>.Name | S3 fn-rule | f |
| Victim's defer argument captured by the caller | zrealm_home_defer_arg | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | l |
| Victim panics after the caller's write attempt | zrealm_home_panic | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | l |
| Caller relays the victim object to a third realm | zrealm_home_third | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | c |
| Read-only relay: caller passes victim object to a library getter | zrealm_home_ro_relay | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | c |
| Registry gate storing a caller function value | zrealm_home_registry | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | g |
| Crossing literal passed to the victim's gate | zrealm_home_crossing_lit | OUT: owner: lit / owner: crossed | R1 | a |
| Known route: pointer | zrealm_home_known_ptr | OUT: owner: ptr | R1 | c |
| Known route: map | zrealm_home_known_map | OUT: owner: map | R1 | c |
| Known route: slice | zrealm_home_known_slice | OUT: owner: table | R1 | c |
| Known route: struct pointer | zrealm_home_known_structptr | OUT: owner: struct | R1 | c |
| Known route: interface | zrealm_home_known_iface | OUT: owner: any | R1 | h |
| Known route: relay table of function values | zrealm_home_known_relay_table | OUT: owner: table | R1 | f |
| Known route: read-only relay picked by the caller | zrealm_home_known_ro_relay | OUT: owner: picked | R1 | c |
| Victim iterates its own objects and calls its own library | zrealm_home_lib_iterate | OUT: owner: kick | R1 | c |
| Victim library handle takes the object and writes | zrealm_home_libhandle | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,0)>.Field | R1 | c |
| Handle to a crossing gate that writes | zrealm_handle_crossing | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | a |
| Handle to a gate obtained generically (`func Gate(cur realm)`) | zrealm_handle_crossing_ref | OUT: set | R1 | a |
| Handle to a plain function observed, then called | zrealm_handle_foreign | OUT: pkgpath: gno.land/r/tests/vm/cbvictim / call: ... read-only | S3 fn-rule | f |
| Own handle to own gate | zrealm_handle_gate_own | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | a |
| Handle stored in a closure | zrealm_handle_inclosure | OUT: closure: ... read-only: Owner<~VPBlock(3,0)> / closure var: | S3 fn-rule | f, g |
| Handle stored in an init-time variable | zrealm_handle_initvar | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)> | S3 fn-rule | f |
| Handle to a library gate that writes same-package state | zrealm_handle_lib_gate | OUT: lib: pwnd-same-pkg (write lands) | R1 | c |
| Handle to a library function writing a foreign object | zrealm_handle_lib_ro | ERR: cannot modify object owned by another realm from library code: x<VPBlock(1,0)>.Field | R1 | c |
| Handle to a method expression over the victim | zrealm_handle_mexpr | OUT: read-only: ... read-only: t<VPBlock(1,0)>.Name | S3 fn-rule | f |
| hiv: real realm handler reached through a forged value | zrealm_hiv_real_rejected | ERR: cannot modify object owned by another realm from library code: c<~VPBlock(1,1)> | R1 | c |
| hiv: unreal (borrowed) realm value tried as a handle | zrealm_hiv_unreal_borrow | OUT: stopped= true | R1 | MISSING |

## 11. Subrealms (`zrealm_sub_*`)

| Scenario | Fixtures | Outcome | Rule | Pattern |
|---|---|---|---|---|
| Sub-realm token printed (`pkgpath#sub/42`) and used | zrealm_sub0 | OUT: gno.land/r/demo/subrealm#dao/42 / true | R1 | MISSING |
| Crossing call from a sub-realm; previous is the sub-token | zrealm_sub1 | OUT: prev.PkgPath: gno.land/r/demo/subrealm#dao/1 | R1 | MISSING |
| `Sub(path)` derive and compare with the token | zrealm_sub2 | OUT: gno.land/r/demo/subrealm dao/42 true | R1 | MISSING |
| Sub-realm as banker scope | zrealm_sub_banker | OUT: sub balance: 600/gno.land/r/demo/subrealm:gold | R1 | MISSING |
| Invalid subpaths rejected by `Sub` | zrealm_sub_derive_reject | OUT: Sub: subpath cannot be empty | R1 | MISSING |
| Sub of a realm the caller does not operate | zrealm_sub_foreign | ERR: Sub: caller is not operating in the receiver's namespace | R1 | MISSING |
| Cross-call into a sub-realm, previous carries subpath | zrealm_sub_foreign_present | OUT: observed prev.Subpath: dao/9 | R1 | MISSING |
| Nested `Sub` on a synthesized receiver | zrealm_sub_nested | OUT: recovered: Sub: receiver pkgpath is already synthesized or invalid | R1 | MISSING |
| Persist a sub-realm token | zrealm_sub_persist | ERR: cannot persist realm value: realm values are ephemeral and tied to a call frame | R1 | MISSING |
| Stale sub-realm token handed to a crossing fn | zrealm_sub_stale | OUT: recovered: Sub: receiver is not the live cur (stale, sibling, or a sub-token) | R1 | MISSING |
| Cross with a stale sub-token | zrealm_sub_stale_cross | ERR: cross: rlm is not the current cur (stale capture or sibling frame) | R1 | MISSING |
| Validation of subpaths by the host realm | zrealm_sub_validation | OUT: rejected: Sub: subpath cannot be empty | R1 | MISSING |

## 12. Objects, views, defers, methods, imports and reference shapes

| Scenario | Fixtures | Outcome | Rule | Pattern |
|---|---|---|---|---|
| Borrow an object across realms via a method with a pointer param | zrealm_borrow0 | OUT: 101 / CodeRealm{...gno.land/r/borrow} | R1, R2 | b |
| Borrowed object's fields mutated through a passthrough realm | zrealm_borrow0a | OUT: s.A: 101 / current realm CodeRealm{...} | R1, R2 | b |
| Borrow with `PrintRealms(_ int, rlm realm)` threading | zrealm_borrow1 | OUT: 101 / CodeRealm{...gno.land/r/borrow} | R1 | a |
| Borrow with a crossing receiver that also prints | zrealm_borrow1a, zrealm_borrow1b | OUT: 101 / CodeRealm{...} | R1 | a, j |
| Borrow through a plain (non-crossing) method | zrealm_borrow2 | OUT: 101 / UserRealm{...} | R1 | b |
| `cross` used inside a composite value (not a call) | zrealm_borrow2a | ERR: use of builtin cross not in function call | R1 | a |
| View callback passed to a library that calls it while the owner is current | zrealm_view_callback | OUT: view callback: cannot modify object owned by another realm... | R1 | e |
| Default view: reader gets a tree handle, write refused | zrealm_view_default | OUT: read: v0 then ERR: cannot modify object owned by another realm | R1 | b |
| Handle that the owner itself operates | zrealm_view_handle | OUT: owner sees: h1 | R1 | b |
| Handle not granted write permission (ingress) | zrealm_view_ingress | OUT: ungranted write by holder: cannot modify object... | R1 | b |
| Owner's method value passed as an argument writes the owner's object | zrealm_view_method_value_arg | OUT: owner's arg: cannot directly modify readonly tainted object | R1 | f |
| Defer with a method call copies the receiver value | zrealm_defer_method0, zrealm_defer_method1 | OUT: before defer fires, c.N = 0 | R2 | d, l |
| Defer with method call arguments evaluated early | zrealm_defer_method_args | OUT: about to return; defers will run LIFO | R2 | d, l |
| Defer mixing a bare function and a method on the same value | zrealm_defer_method_mixed | OUT: function: bare / method appended; tag now: x-mid | R1 | l |
| Deferred method on a stored (multihop) object | zrealm_defer_method_multihop | OUT: stored.N before defer fires = 0 | R2 | d, l |
| Deferred method on a nil receiver | zrealm_defer_method_nil | OUT: nil greeted | R1 | l |
| Deferred method that recovers an outer panic | zrealm_defer_method_outer_recover | OUT: Closed before defer fires: false | R1 | l |
| Deferred method expression | zrealm_method_expr_defer | OUT: body done | S3 fn-rule | f, l |
| Method expression invoked inline | zrealm_method_expr_inline | OUT: should not print (no output expected) | S3 fn-rule | f |
| Unexported method expression from another package | zrealm_method_expr_unexported | ERR: cannot access unexported method gno.land/p/demo/tests/v0.clearField from gno.land/r/ | R6 | h |
| Interface method receiver from another package | zrealm_method_iface | OUT: 7 | R6 | h |
| Import a package closure and call it across realms | zrealm_import_closure_root | ERR: cannot copy to readonly tainted slice | R1, R2 | d |
| Import and use a method value from another package | zrealm_import_mv, zrealm_import_mv_caller | ERR: cannot copy to readonly tainted slice | R2 | d |
| Nested import of a realm helper | zrealm_import_nested | ERR: cannot copy to readonly tainted slice | R2 | d |
| Stdlib sort.Interface implemented by another realm's slice | zrealm_import_std_in | OUT: sorted: 1 2 3 | R1 | c |
| Stdlib writing a foreign slice out-parameter | zrealm_import_std_out | OUT: control: efgh then ERR: cannot copy to readonly tainted slice | R2 | d |
| Reference: pointer to own struct returned and written | zrealm_ref0 | OUT: 2 | R2 | b |
| Reference: own struct value printed with address | zrealm_ref1 | OUT: (struct{(20 int)} gno.land/r/test.MyStruct) | R2 | b |
| Reference: pointer to interface{} | zrealm_ref2, zrealm_ref4 | OUT: *interface {} | R2 | b |
| Reference: interface holding a string | zrealm_ref3 | OUT: hello | R2 | h |
| Reference: write through `*interface{}` from a library fn | zrealm_ref5 | OUT: changed | R2 | d |
| Reference: interface replaced through a pointer | zrealm_ref6, zrealm_ref7 | OUT: hello / replaced | R2 | d, h |
| Reference: tree with a callback during insert | zrealm_ref8 | OUT: 10 | S3 fn-rule | e |
| Reference: tree built in init | zrealm_ref9 | OUT: 6 | R1 | a |
| Map: print map of own type | zrealm_map0 | OUT: map{("foobar" string):(1 int)} | R2 | d |
| Map: caller-realm map written from library code, refused | zrealm_map1 | OUT: after: m[admin] = 0 (write refused, unchanged) | R1 | c |
| Map: attacker key added to a foreign map | zrealm_map2 | OUT: len(m) = 2 / attacker key exists: true | R1 | c |
| Map: foreign map entries read/write in init | zrealm_map3 | OUT: admin exists: false / user1: 500 | R1 | c |
| Map: own map grown in init and read back | zrealm_map4, zrealm_map4a, zrealm_map5 | OUT: len: 3 / a: 1 / b: 2 | R1 | a |
| Map: nil map range | zrealm_mapnil | OUT: ok, len: 0 | R2 | d |
| Slice: own slice literal passed to a gate | zrealm_slice0 | OUT: s[0]: 1 / s[1]: 2 / s[2]: 3 | R2 | d |
| Slice: all elements set through a pointer to the array | zrealm_slice1 | OUT: s[0]: 99 / s[1]: 99 / s[2]: 99 | R2 | d |
| Slice: copy semantics on range | zrealm_slice2 | OUT: copied: 3 / 0 99 / 1 88 | R2 | d |
| Slice: len/idx range over own slice | zrealm_slice3 | OUT: len: 3 / 0 10 / 1 20 | R2 | d |
| AVL tree (stdlib) used as a realm store | zrealm_avl0, zrealm_avl1, zrealm_avl2 | OUT: false 2 / false 3 | R1 | c |
| Banker: `GetCoin` with an uppercase denom | zrealm_banker_getcoin_denom | OUT: 0 then ERR: invalid denom: UPPERCASE | R1 | k |
| Banker: `NewBanker` over `cur.Previous()` | zrealm_banker_previous | ERR: banker can only be instantiated for the current realm | R1 | MISSING |
| Banker: `TotalCoin` with an uppercase denom | zrealm_banker_totalcoin_denom | OUT: 0 then ERR: invalid denom: UPPERCASE | R1 | k |
| Const across realms | zrealm_const | OUT: 20 | R1 | d |
| databyte: byte slice storage | zrealm_databyte0 | OUT: ABbb | R1 | d |
| init-time cycle of two objects (unescaped) | zrealm_cycle_unescaped | OUT: ok | R1 | a |
| Cycle of objects built later through crossing calls | zrealm_cycle_unescaped_later | OUT: true / 42 / 9 42 | R1 | a |
| Example realm used as a fixture | zrealm_example | OUT: # [plot] First Plot | R1 | a |
| FuncPkgPath: top-level, literal, other-realm, method | zrealm_funcpkgpath | OUT: other realm top-level: gno.land/r/tests/vm/funcpkg | S3 fn-rule | f |
| FuncPkgPath: method value of an interface | zrealm_funcpkgpath_iface, zrealm_funcpkgpath_iface_big | ERR: runtime.FuncPkgPath: method value of an interface has no fixed declaration | R6 | f, h |
| FuncPkgPath: interface swapped after the value was taken | zrealm_funcpkgpath_iface_swap | OUT: lazy before swap panicked: runtime.FuncPkgPath: method value of an interface... | S3 fn-rule | f |
| FuncPkgPath: not a function / untyped nil | zrealm_funcpkgpath_notfunc, zrealm_funcpkgpath_untypednil | ERR: runtime.FuncPkgPath: not a function | S3 fn-rule | f |
| init-time cur printing | zrealm_initctx | OUT: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm | R1 | MISSING |
| `cur.IsCurrent()` inside a gate | zrealm_iscurrent | OUT: cur.IsCurrent(): true / cur.Previous().IsCurrent(): false | R1 | i |
| List byte append into a foreign slice | zrealm_listbyte_append0 | OUT: 90 | R2 | d |
| model token: library ledger moved by owner, refused by holder | zrealm_model_token | OUT: bob: 10 then ERR: cannot modify object owned by another realm: l<...>.bal[from] | R1 | c |
| natbind: natives bound in init and after | zrealm_natbind0 | OUT: 123 / dev | R1 | k |
| natbind: package path from a native | zrealm_natbind1 | OUT: gno.land | R1 | k |
| p-attacker via interface: victim write refused, state intact | zrealm_p_attacker_via_iface_filetest | OUT: caller Touch: cannot modify object... / after Touch A= 1 | R1 | h |
| p-convert: converting a readonly foreign value | zrealm_p_convert_readonly_ok_filetest | ERR: cannot copy gno.land/r/tests/vm.Foo outside realm gno.land/r/tests/vm (use a pointer) | R2 | d |
| Panic declared by a non-crossing realm method | zrealm_panic | ERR: panic | R1 | l |
| `realm` value stored in the caller's state after the call | zrealm_r_stored_in_caller_filetest | OUT: B: LocalStruct{123} / CurrentRealm after return: gno.land/r/caller | R2 | b |
| Readonly receiver: direct pointer write | zrealm_readonly_recv_direct | ERR: cannot directly modify readonly tainted object: p<VPBlock(1,1)>.Field | R1 | b |
| Readonly receiver: mutate a package (not its realm state) | zrealm_readonly_recv_method | ERR: cannot mutate gno.land/p/demo/tests/v0: package is immutable post-init | R1 | MISSING |
| Readonly receiver: value-receiver method writes its copy | zrealm_readonly_recv_method_val | OUT: after ModifyVal; v.Field = init (copy written, original intact) | R2 | d |
| Runtime: nil pointer dereference inside a realm | zrealm_runtimeerror | OUT: runtime error: nil pointer dereference | R1 | l |
| Sealed realm: forged realm type via embedding | zrealm_seal_embed | ERR: gno.land/r/crossrealm.fooRealm does not implement .uverse.realm (missing method .seal) | R6 | MISSING |
| Same forged type through a pointer embed | zrealm_seal_embed_ptr | ERR: gno.land/r/crossrealm.fooRealm does not implement .uverse.realm (missing method .seal) | R6 | MISSING |
| Forged realm type asserted at runtime | zrealm_seal_embed_runtime | OUT: assertion ok: false | R1 | MISSING |
| Forged realm struct implementing realm methods | zrealm_seal_realm | ERR: gno.land/r/crossrealm.fakeRealm does not implement .uverse.realm (missing method .seal) | R6 | MISSING |
| std: cur address / pkgpath from a plain fn | zrealm_std0, zrealm_std3, zrealm_std5 | OUT: main gno.land/r/std | R1 | i |
| std: crossing call between std-ish realms | zrealm_std4 | OUT: test1 gno.land/r/tests/vm | R1 | a |
| stdlib global write from a realm | zrealm_stdlib_global_write | OUT: 1 / 2 | R1 | MISSING |
| stdlib ref (hex/b64 helper holds a value) | zrealm_stdlib_ref | OUT: 42 | R1 | b |
| b64 encode into a destination slice | zrealm_b64_encode_dst | OUT: aGk= | R2 | d |
| strconv Ftoa on a realm receiver | zrealm_strconv_ftoa_rcaller | OUT: ok, no panic | R1 | c |
| Stringer list built in init(cur realm) | zrealm_tests0 | OUT: 0: &FooStringer{one} | R1 | MISSING |
| Interface implemented and appended in a realm | zrealm_tests1 | OUT: ok | R1 | h |
| Test utils address helpers | zrealm_testutils0 | OUT: g1w3jhxap3ta047h6... | R1 | k |
| v2: `new` of a foreign declared type | zrealm_v2_construction_new_fail | ERR: cannot allocate gno.land/r/tests/vm/crossrealm.LocalStruct in realm gno.land/r/bar | R2 | d |
| v2: composite literal of a foreign declared type | zrealm_v2_construction_composite_fail | ERR: cannot allocate gno.land/r/tests/vm/crossrealm.LocalStruct in realm gno.land/r/bar | R2 | d |
| v2: mutate a foreign byte slice | zrealm_v2_foreign_mutate_unreal_bytes | ERR: gno.land/r/tests/vm/crossrealm_d's code was entered without crossing: bz<VPBlock(1,0)>[0] | R1 | b |
| value_ban: copy a foreign declared value after a gate ran | zrealm_value_ban | OUT: 1 2 then ERR: cannot copy gno.land/r/tests/vm/ctw.Local outside realm (use a pointer) | R2 | d |
| value_ban2: allocate a foreign coupon type in a range | zrealm_value_ban2 | ERR: cannot allocate gno.land/r/tests/vm/valueban.Coupon in realm gno.land/r/valueban2 | R2 | d |

## 13. Example fixtures used by the filetests (`examples/gno.land/`)

| Scenario | Fixtures | Outcome | Rule | Pattern |
|---|---|---|---|---|
| Gate writes its counter and a stored object; unexported helper bumps | r/tests/vm/ctw/ctw.gno | OUT: 3 2 (via zrealm_ctw_cross) | R1 | a |
| Library table runs caller callbacks (`Run`, `Apply`, `Each`) | p/demo/tests/libtable/v0/libtable.gno, r/tests/vm/ctw/ctw.gno | OUT: refused when the callback writes its own realm | R1 | c, e |
| Library table dispatches its own writer `Inc` over a box | p/demo/tests/libtable/v0, r/tests/vm/ctw/ctw.gno | OUT: boxes: 6 (via zrealm_ctw_owncb) | R1 | c, f |
| Victim's `Panics` value taken and called inside a gate | p/demo/tests/libtable/v0, r/tests/vm/ctw | OUT: recovered: boom (via zrealm_ctw_recover) | S3 fn-rule | f, l |
| Owner realm stores a table of its own writers | r/tests/vm/cbvictim/v.gno | OUT: owner: own-table (via zrealm_home_own_table) | R1 | a, f |
| `RunAll` gate runs caller-supplied values inside the gate | r/tests/vm/cbvictim/v.gno | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it | S3 fn-rule | m |
| `RunS` gate runs one picked value | r/tests/vm/cbvictim/v.gno | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it | S3 fn-rule | a |
| `RelayFrom` non-crossing relay of a caller value | r/tests/vm/cbvictim/v.gno | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it | S3 fn-rule | m |
| Caller realm hands callbacks and containers to the victim | r/tests/vm/cbother/other.gno | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it | S3 fn-rule | m |
| Closure capture and write back into the capturing realm | r/tests/vm/closurevictim/closurevictim.gno, closurehelper | OUT: write refused once the owner is not current | R1 | e |
| Realm with a crossing method and a plain method on the same type | r/tests/vm/crossrealm/crossrealm.gno | OUT: hello Foo gno.land/r/crossrealm / hello D ... | R1, R6 | j |
| Realm that sets an object owned by another realm through a gate | r/tests/vm/crossrealm/switchrealm.gno, crossrealm_b | OUT: SetObject writes 123 | R1, R2 | a, j |
| Realm helper `d` that panics after a refused write | r/tests/vm/crossrealm_d | OUT: panic caught: ... read-only | R1 | l |
| Realm helper `e` with a recover-based attack | r/tests/vm/crossrealm_e | OUT: exception caught: ... read-only / owner unchanged | R1 | l |
| Realm helper `f` used by borrow fixtures | r/tests/vm/crossrealm_f | OUT: 101 (borrow read/write) | R1, R2 | b |
| Chain realms: cur/previous pkgpath at each hop | r/tests/vm/curchain_a, curchain_b | OUT: A prev.PkgPath: gno.land/r/curchaintest | R1 | i |
| Entry-point signature variants (gate named cur/other first arg) | r/tests/vm/entries/entries.gno | OUT: gate variants accepted or refused at type-check | R1 | a |
| Entry closure library (`call(cur, f)`) | r/tests/vm/entryclosure/entryclosure.gno, p/demo/tests/entryclosure/v0 | OUT: p closure: cannot modify object... | R1 | e |
| Intrusion attempts assembled as one exploit fixture | r/tests/vm/exploit.gno | OUT: writes refused, victim state intact | R1 | c |
| FuncPkgPath fixtures: declared fn, method, literal, cross-pkg | r/tests/vm/funcpkg/funcpkg.gno, p/demo/tests/funcpkg/v0 | OUT: other realm top-level: gno.land/r/tests/vm/funcpkg | S3 fn-rule | f |
| Declared struct/error types (`T`, `E`, `H`) used by R2 tests | r/tests/vm/fvx/fvx.gno | ERR: cannot copy ... outside realm (use a pointer) | R2 | d |
| Interface attack values and victim interface | r/tests/vm/ifacevictim/v.gno, p/demo/tests/ifaceattack/a.gno | ERR: cannot modify object owned by another realm from library code | R1, R6 | h |
| Interface importer realm | r/tests/vm/ifaceimporter/v.gno | ERR: cannot modify object owned by another realm | R1, R6 | h |
| Interface wrapper use realm | r/tests/vm/ifacewrapuse/u.gno, p/demo/tests/ifacewrap/w.gno | OUT: iface / ifacev | R6 | h |
| Import victim: readonly slice written through an imported helper | r/tests/vm/impvictim/v.gno, impvictim/control/v.gno | ERR: cannot copy to readonly tainted slice | R2 | d |
| Inner owner realm writes a nested foreign object | r/tests/vm/innerowner/innerowner.gno | ERR: cannot modify object owned by another realm | R1 | c |
| Interface fixture realm (Stringer, Get) | r/tests/vm/interfaces.gno | OUT: 7 | R6 | h |
| rdata victim realm (Immutable, Bare, stored pointers) | r/tests/vm/launderrvictim/launderrvictim.gno | ERR: cannot directly modify readonly tainted object | R1, R2 | b, d |
| Laundering victim realm (Object, pointer/value fields) | r/tests/vm/laundervictim/laundervictim.gno | ERR: cannot modify object owned by another realm | R1 | c |
| Model token realm storing a library ledger | r/tests/vm/modeltoken/t.gno, p/demo/tests/modelledger/l.gno | OUT: bob: 10 then ERR: cannot modify ... .bal[from] | R1 | c |
| Persisted closure cap fixture | r/tests/vm/persistedcap/persistedcap.gno | ERR: cannot modify object owned by another realm from library code (call a function of the owning realm): c<~VPBlock(1,1)> | S3 fn-rule | g |
| Stored function value victim realm (`Hook Fn`) | r/tests/vm/pfuncvictim/v.gno | ERR: cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,0)>.Field | S3 fn-rule | g |
| Realm composite literal fixture | r/tests/vm/realm_compositelit.gno | OUT: literal built and printed | R2 | d |
| Realm method over 38 derived types | r/tests/vm/realm_method38d.gno | OUT: method set resolved per type | R6 | j |
| Registry caller-identity fixtures (alias, noop, spoof, approve, from, user) | r/tests/vm/regalias, regnoop, regspoof, regspoofapprove, regspooffrom, regspoofuser, regzero | OUT: caller identity checks pass or refuse spoofing | R1 | i |
| Sub-realm host fixture | r/tests/vm/subhost/subhost.gno | OUT: observed prev.PkgPath: ...#dao/5 | R1 | MISSING |
| Sub-test table realm | r/tests/vm/subtests/subtests.gno, p/demo/tests/subtests/v0 | OUT: subtest cases pass | R1 | a |
| Test20 fixture realm | r/tests/vm/test20/test20.gno | OUT: fixture-specific prints | R1 | a |
| Test helper foo realm | r/tests/vm/tests_foo/foo.gno | OUT: helper used by tests.gno | R1 | a |
| Generic realm test fixtures (tests.gno, tests_test.gno) | r/tests/vm/tests.gno, tests_test.gno, nestedpkg_test.gno | OUT: package tests pass | R1 | a |
| value ban victim realm | r/tests/vm/valueban/value.gno | ERR: cannot copy ... outside realm (use a pointer) | R2 | d |
| Variadic crossing function fixture | r/tests/vm/variadic/main.gno | OUT: variadic gate called with cross | R1 | a |
| Default view realm (`viewdefault`) | r/tests/vm/viewdefault/viewdefault.gno, p/demo/tests/viewgrant/v0 | OUT: read: v0 then ERR: cannot modify object | R1 | b |
| In-repo filetests exercising tests/vm realms (z0–z4) | r/tests/vm/filetests/z0_filetest.gno ... z0–z3_filetest.gno, z4_filetest.gno | OUT: per-filetest goldens | R1 | a |
| Callback attack library | p/demo/tests/cbattack/a.gno | ERR: cannot modify object owned by another realm | R1 | e |
| Launder attack library (`EvilWrite`) | p/demo/tests/launderattack/v0/launderattack.gno | OUT: pwnd-via-apply where the maker rule allows it | R1 | f |
| Launder package: `Immutable`, `Object`, `Bare` types | p/demo/tests/launderpkg/v0/launderpkg.gno | ERR: cannot modify object owned by another realm | R1, R2 | c, d |
| Nested closure library | p/demo/tests/nestclosure/v0/nestclosure.gno | OUT: lit: ... entered without crossing | R1 | e |
| Closure cap library | p/demo/tests/p_closurecap/v0/p_closurecap.gno | OUT: closure capture write refused | R1 | e |
| Cross-realm package helper | p/demo/tests/p_crossrealm/v0/p_crossrealm.gno | OUT: cross into the helper's realm | R1 | a |
| Box getter library | p/demo/tests/pboxget/p.gno | OUT: box returned, write refused | R1 | b |
| Callback binding library (`Fn`, `Apply`) | p/demo/tests/pfuncbind/p.gno | ERR: cannot modify object owned by another realm | R4 fn-rule | f |
| Test realm object and writers (`TestRealmObject`, `Writers`) | p/demo/tests/v0/tests.gno, v0/tests_test.gno | OUT: (struct{("modified" string)} ...TestRealmObject) | R1 | a, f |

## 14. Integration txtars (`gno.land/pkg/integration/testdata/`)

| Scenario | Fixtures | Outcome | Rule | Pattern |
|---|---|---|---|---|
| MsgCall into a crossing gate from a user account | maketx_call_pure.txtar, interrealm_mix_call.txtar | ERR: gno.land/r/test/borrowrealm's code was entered without crossing and is read-only | R1 | k |
| MsgCall of a non-crossing mutator | interrealm_mix_call.txtar | ERR: borrowrealm's code was entered without crossing and is read-only | R1 | k |
| Run script (`maketx run`) calls a realm gate | interrealm_mix_run.txtar, maketx_run.txtar, maketx_run_send.txtar | OUT: OK! / non-crossing mutation refused | R1 | k |
| Run script mutates a realm without crossing | interrealm_mix_run.txtar | ERR: borrowrealm's code was entered without crossing and is read-only | R1 | k |
| Run script sends coins and reads `OriginSend` | maketx_run_send.txtar | OUT: send: 42ugnot | R1 | k |
| MsgCall with a send envelope | maketx_call_send.txtar | OUT: ("send: 42ugnot") | R1 | k |
| Array-argument decoding for a realm call | maketx_call_array_len.txtar | OUT: ("32" string) then ERR: array length mismatch | R1 | k |
| Run-script realm compared with a deployed realm (subrealm parity) | subrealm_run_parity.txtar | OUT: observed prev.Subpath: dao/5 / prev.IsUserCall: false | R1 | MISSING |
| Two-realm crossing chains (final corpus) | interrealm_final.txtar | ERR: gno.land/r/test/bob's code was entered without crossing | R1 | a |
| Adopted functions from a run script | interrealm_final_omarsy.txtar | ERR: cannot modify object owned by another realm from library code | S3 fn-rule | g |
| `cur` binding through a wrapper realm | interrealm_cur_binding.txtar | OUT: ok=true | R1 | i |
| Read-only access across realms (OK corpus) | interrealm_read_only.txtar, interrealm_read_only_declared_type.txtar, interrealm_v2.txtar | OUT: OK! | R1 | b |
| Assign a foreign object's field, recover, then retry | crossrealm_assign_recover.txtar | ERR: unauthorized | R1 | l |
| Home-realm handler passed arrays/pointers through a caller | home_alias_arr.txtar, home_route_arr.txtar, home_route_arrarr.txtar, home_route_ptr.txtar | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it | S3 fn-rule | m |
| Home handler with an object deleted between transactions | home_deleted_ref.txtar | ERR: unexpected object with id | R1 | MISSING |
| Home handler kept across two transactions | home_kept_twotx.txtar | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it | S3 fn-rule | m |
| Nested load of a home object | home_nested_load.txtar, home_route_outer.txtar | ERR: cbvictim's code is read-only here: a function value writes only in the crossing call that made it | S3 fn-rule | m |
| Save a handler into a home realm | home_save.txtar | ERR: gno.land/r/demo/hv's code is read-only here: a function value writes only in the crossing call that made it | S3 fn-rule | g |
| Stored handler fired later | home_stored_later.txtar | OUT: fired / all | S3 fn-rule | g |
| Unloaded home object append/copy | home_unloaded.txtar | ERR: gno.land/r/demo/hv's code is read-only here: a function value writes only in the crossing call that made it | S3 fn-rule | m |
| Adopt a function value from a run script | maker_adopt.txtar, maker_adopt_iface.txtar | ERR: cannot modify object owned by another realm from library code | S3 fn-rule | g |
| Get a method value from a victim and call it | maker_getmv.txtar | ERR: gno.land/r/demo/zvic's code is read-only here: a function value writes only in the crossing call that made it | S3 fn-rule | f |
| Named function value across a node restart | maker_named_restart.txtar | OUT: "named" string / "value" string | S3 fn-rule | g |
| Stored function value fired and then called directly | maker_stored.txtar | ERR: gno.land/r/demo/mka's code is read-only here: a function value writes only in the crossing call that made it | S3 fn-rule | g |
| Anonymous own-function routes | fv_anon_own.txtar, fv_anon_own_private.txtar, fv_anon_own_routes.txtar | OUT: new / again / later / after | S3 fn-rule | f |
| Lazy function value copy refused | fv_copy_lazy.txtar | ERR: cannot copy gno.land/r/test/fva.T outside realm (use a pointer) | R2 | d |
| Named function value across restart (pfunc) | pfunc_named_restart.txtar | OUT: "pwnd-via-apply" string | S3 fn-rule | g |
| Stored interface pointer used from a run script | iface_storedptr_2tx.txtar | ERR: cannot modify object owned by another realm from library code | R1, R6 | h |
| Unstamped method value laundered | iface_unstamped_mv.txtar | ERR: cannot modify object owned by another realm | R1, R6 | h |
| Default view adopted, not shared | view_default_adopted_not_shared.txtar | ERR: call a function of the owning realm | R1 | b |
| Method value of an adopted view | view_method_value_adopted.txtar | ERR: cannot modify object owned by another realm from library code | S3 fn-rule | g |
| Stored view reference refused | view_stored_ref_refused.txtar | OUT: OK! (write refused) | R1 | b |
| Object pointer stored, then used across realms | object_pointer.txtar, object_pointer_2.txtar, object_pointer_pure.txtar | OUT: 42 int / ("reply from ext realm" string) / ERR: package is immutable post-init | R2 | d |
| PreviousRealm assertions from a realm | prevrealm.txtar | OUT: test1_user_addr / RFOO_USER_ADDR | R1 | i |
| Render calls a crossing function | render_crossing.txtar | OUT: Render cur.Previous().PkgPath() = "" | R1 | a |
| Origin call assertions | assertorigincall.txtar, assertorigincall_alias.txtar | ERR: invalid non-origin call | R1 | i |
| Allowance sender handoff / recovery | allowancesender_handoff.txtar, allowancesender_recover.txtar, allowancesender_outer_recover.txtar, allowancesender_persistence.txtar, allowancesender_diag.txtar | ERR: closed / cap exceeded / cannot modify | R1 | MISSING |
| Payable callee reads the send envelope | payable_callee_read.txtar, payable_observed_send.txtar | ERR: never read the send-envelope | R1 | k |
| OriginSend lifetime | banker_originsend_lifetime.txtar, banker_persistence.txtar, banker_realm_embed.txtar, banker_security.txtar | OUT: banker-scoped balances behave per call | R1 | MISSING |
| Event callback fired from a realm | event_callback.txtar, event_defer_callback_loop.txtar | OUT: EVENTS: [{"type":"foo"...}] | R1 | MISSING |
| Function value handle across a node restart | fnvalue_handle_restart.txtar | ERR: gno.land/r/test/hr's code is read-only here: a function value writes only in the crossing call that made it | S3 fn-rule | g |
| Ephemeral callback launder (call and run) | ephemeral_callback_launder.txtar, ephemeral_method_callback_launder.txtar | ERR: cannot modify object owned by another realm | R1 | e |
| Caller-teller home pattern (grc20/grc721) | grc20_callerteller_home.txtar, grc721_callerteller_home.txtar, grc20_registry*.txtar | OUT: registry emits and wrappers forward calls | R1 | MISSING |
| Pointer map key across realms | ptr_mapkey.txtar, mapkey.txtar, map_delete.txtar, map_storage.txtar | OUT: 5 int | R2 | d |
| Interface method persistence across restart | method_iface_cyclic_persist.txtar, method_iface_lazy_persist.txtar, method_iface_shadow_persist.txtar, method_nil_value_persist.txtar | OUT: restarted state restored | R6 | h |
| FuncPkgPath over a stored value | funcpkgpath_stored.txtar | OUT: path of the stored function | S3 fn-rule | f |
| Realm banker issued coin denom | realm_banker_issued_coin_denom.txtar | OUT: issued denom reported | R1 | MISSING |
| Reference cycle guard across restart | realm_cycle_guard_restart.txtar, cycle fixtures | OUT: cycle survives restart | R1 | a |

## 15. Real realms in `examples/gno.land/r` (non-test)

| Scenario | Realm (representative file) | Outcome | Rule | Pattern |
|---|---|---|---|---|
| Gate + `cur.Previous().IsUserCall()` + `OriginSend` for a paid call | demo/disperse/disperse.gno | OUT: paid disperse; non-EOA caller refused | R1 | a, i, k |
| Cross-realm token transfer with `cross(cur)` | demo/disperse/disperse.gno (tokens.TransferFrom) | OUT: transfer lands | R1 | a |
| Paid swap gate: only direct user call accepted | demo/defi/atomicswap/atomicswap.gno | OUT: swap completes; proxy callers refused | R1 | i, k |
| Stored crossing function value invoked with `cross(rlm)` | demo/defi/atomicswap/swap.gno (`s.sendFn(cross(rlm), ...)`) | OUT: payout dispatched per side | R1 | f |
| Registry registration with `cross(cur)` | demo/defi/foo20/foo20.gno, gnoland/wugnot/wugnot.gno | OUT: token registered in grc20reg | R1 | a |
| Factory registers a token on behalf of the caller | demo/defi/grc20factory/grc20factory.gno | OUT: token registered | R1 | a |
| OriginSend payment check on a wrapped coin | gnoland/wugnot/wugnot.gno | OUT: wrap/unwrap with sent coins | R1 | k |
| Profile realm: gate writes own state, render reads | demo/profile/profile.gno | OUT: profile set/read | R1 | a, b |
| Boards: per-action `IsUserCall` guards, fee via OriginSend | archive/boards/public.gno | OUT: posts accepted; fee required | R1 | i, k |
| Boards2: gates over several files, `IsUser` caveat documented | gnoland/boards2/v0/public_invite.gno | OUT: invite/reply writes land | R1 | i |
| DAO: proposals executed by crossing into the impl realm | gov/dao/impl/v0/govdao.gno | OUT: proposal executes | R1 | a |
| DAO: memberstore called with `cross(cur)` from impl code | gov/dao/impl/v0/impl.gno, prop_requests.gno | OUT: membership updated | R1 | a |
| DAO: non-crossing helper threads `rlm` into `cross(rlm)` reads | gov/dao/impl/v0/types.gno | OUT: membership read | R1 | a |
| DAO: loader/init set up tiers through crossing calls | gov/dao/loader/v0/loader.gno, gov/dao/init/v0/init.gno | OUT: tiers set, impl updated | R1 | a |
| DAO: treasury payment from a proposal | gov/dao/impl/v0/prop_requests.gno (`treasury.Send(cross(cur), ...)`) | OUT: payment sent | R1 | a |
| DAO: memberstore stores members (sub-realm-ish state) | gov/dao/memberstore/v0 | OUT: tiers/members persisted | R1 | b |
| Valopers: `IsUserCall` anti-squat + OriginSend fee | gnops/valopers/valopers.gno | OUT: register/rotate with fee | R1 | i, k |
| Valopers: notify validators through a crossing call | gnops/valopers/valopers.gno (`validators.NotifyValoperChanged(cross(cur), ...)`) | OUT: validator notified | R1 | a |
| Valopers admin: same-package cross needs no trick | gnops/valopers/admin.gno | OUT: admin op lands | R1 | a |
| Valopers proposal realm: request objects built with `cross(cur)` | gnops/valopers/proposal/proposal.gno | OUT: proposal request created | R1 | a |
| Params: cross-realm authority changes | sys/params/*.gno | OUT: param set; unauthorized caller refused | R1 | i |
| Validators: gates for valset changes called by other realms | sys/validators/v0, v2 | OUT: valset op applied | R1 | a |
| Namereg: name resolution gates, per-caller checks | sys/namereg/v0 | OUT: name registered; spoof refused | R1 | i |
| Users: controller add/remove through gates | sys/users, sys/users/init | OUT: controllers updated | R1 | a |
| CLA: membership writes gated | sys/cla | OUT: CLA accepted | R1 | a, i |
| Txfees: fee collector set through a gate | sys/txfees | OUT: collector set | R1 | a |
| Names: pause/unpause by governance | sys/names | OUT: pause applied | R1 | a |
| GRC20 registry: registration forwarded across realms | nt/grc20reg/v0/grc20reg.gno | OUT: registry entry added | R1 | a |
| Commondao: proposals, treasury, council crossing calls | nt/commondao/v0/public.gno, proposal_treasury.gno, reentrancy.gno | OUT: proposal executed; reentrancy fixture refuses | R1 | a, i |
| Home realm: markdown listing realms | gnoland/home | OUT: rendered listing | R1 | b |
| Blog: posts written from the owner's gate | gnoland/blog | OUT: post published | R1 | a |
| ghverify: identity verification from a crossing call | gnoland/ghverify | OUT: verification recorded | R1 | a |
| Events realm: emits events from a gate | devrels/events | OUT: event emitted | R1 | k |
| Security patterns doc realm (teaching fixture) | docs/security_patterns/security_patterns.gno | OUT: writes gated; PreviousRealm caveat shown | R1 | i |
| Fee receipt realm | gnops/valopers/fee receipt path | OUT: receipt issued | R1 | k |
| Pause/upgrade helpers | sys/* | OUT: ops applied through gates | R1 | a |

## 16. MISSING: scenarios no whitepaper pattern shows

Grouped so the next round can add them either as a new pattern (NP) or as one
more call inside an existing pattern (EC). One representative fixture each.

1. **Sub-realm tokens** (`pkgpath#sub`, `Sub()`, stale sub-tokens, sub-realm
   banker) — representative `zrealm_sub1`; also `zrealm_sub_banker`,
   `zrealm_sub_stale`, `subrealm_run_parity.txtar`, `r/tests/vm/subhost`. NP.
2. **`func init(cur realm)` as a gate** (init sees itself as cur and the user as
   previous; init crossing into other realms/deploy ordering) — `zrealm_cur_init`,
   `zrealm_crossrealm15`, `zrealm_crossrealm25a`, `zrealm_initctx`,
   `zrealm_tests0`, `zrealm_crossrealm38`. NP.
3. **`cur` persistence is a type/state error** ("cannot persist realm value")
   for var, struct field and stored closure — `zrealm_cur_persist_var`; also
   `zrealm_sub_persist`. NP (pattern c stops at lifetime, not storage).
4. **Package immutability** ("cannot mutate gno.land/p/...: package is
   immutable post-init") on a readonly receiver — `zrealm_readonly_recv_method`,
   `object_pointer_pure.txtar`. NP.
5. **Sealed realm type** (a forged `realm` implementation cannot be built:
   "missing method .seal") — `zrealm_seal_embed`, `zrealm_seal_embed_runtime`,
   `zrealm_seal_realm`. NP.
6. **Language-shaped refusals around `cur`**: taking its address, aliasing the
   realm type, naming another parameter `cur`, reassigning/shadowing, range
   assignment — `zrealm_cur_addr`, `zrealm_cur_alias`, `zrealm_cur_reassign`,
   `zrealm_cur_range`. EC (i), as a "binding rules" list.
7. **Deleted language surface kept as fixtures**: `mutable()`, `cross1`,
   `cross` used outside a call — `zrealm_entry_no_mutable`,
   `zrealm_cross1_removed`, `zrealm_borrow2a`. EC (a), as historical notes.
8. **Deferred crossing calls** ("cannot be deferred directly"; legal form is a
   deferred closure) — `zrealm_cross_defer`, `zrealm_cross_deferwrap`. EC (a).
9. **`unsafe` ImportPolicy**: `rust` dot-import absence / which symbols a realm
   may import (`name unsafe not declared`) — `zrealm_crossrealm12`. EC (i).
10. **Stored value deleted between transactions** ("unexpected object with
   id") — `home_deleted_ref.txtar`. NP under pattern x (known limits).
11. **Fn-value laundering through package vars that changes state** — the two
    fixtures whose golden is a successful write via `pdata`:
    `zrealm_launder_pdata_ptr_deref`, `zrealm_launder_pdata_value_copy`. These
    contradict patterns f/g as written (an adopted value writes the victim) and
    must be reconciled before they are cited. NP + flag for S3.
12. **`IsUserRun` / run-script-specific identity** (`maketx run` ephemeral
    realm, `cur.Previous().IsUserRun()`) — `zrealm_cur_classify`,
    `zrealm_cur_parity`, `interrealm_mix_run.txtar`. EC (k), currently only
    `IsUserCall` is documented.
13. **Allowance-sender handoff** (capability granted to a caller's realm, then
    closed/exceeded) — `allowancesender_handoff.txtar`,
    `allowancesender_persistence.txtar`. NP.
14. **Banker scope and lifetime** (`NewBanker` only for the live cur; issued
    denom; persistence across calls) — `zrealm_banker_previous`,
    `banker_originsend_lifetime.txtar`, `realm_banker_issued_coin_denom.txtar`.
    EC (k).
15. **Event emission from a realm/callback** — `event_callback.txtar`,
    `event_defer_callback_loop.txtar`, `devrels/events`. EC (k).
16. **Caller-teller home pattern** (a token realm registers a *teller* realm
    that is the home for holder calls) — `grc20_callerteller_home.txtar`,
    `grc721_callerteller_home.txtar`. NP.
17. **`Render` calling a crossing function** — `render_crossing.txtar`,
    `zrealm_ctw_render`. EC (a) or (k).
18. **hiv: unreal borrowed realm value used as a handle** — `zrealm_hiv_unreal_borrow`
    (golden: `stopped= true`). EC under pattern x.

## 17. Counts

| Metric | Count |
|---|---|
| `zrealm_*` fixtures read | 460 |
| Other filetests touching realms/crossing read | 92 |
| `examples/gno.land/r/tests/vm` fixtures read | 56 |
| `examples/gno.land/p/demo/tests` helpers read | 18 |
| Integration txtars read (of 263; interrealm subset) | 60 |
| Non-test realm files read (realm directories) | 61 (36) |
| Families (tables) | 16 |
| Scenario rows | 594 |
| Covered by a pattern (a–m) | 554 |
| Missing (need a new pattern or extra call) | 40 rows in 18 groups (see §16) |

Notes for the next doc round

- Patterns a, l, i and w carry most of the corpus; pattern x is where the
  home-realm handler fixtures (`zrealm_home_*`) belong today, but they read as
  their own family and may deserve a pattern of their own.
- Everything marked **fn-rule** in §6, §7 and §9 depends on S3 or R4 and
  should be re-checked when either changes; §16 item 11 in particular is a live
  contradiction, not a documentation gap.
- No code or golden was changed for this inventory; it is a documentation-only
  working file and may be dropped before merge.


