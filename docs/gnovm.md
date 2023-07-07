---
state: improvements-needed # The final PoC is working well, but it must evolve performance-wise.
---

# Gnovm

Gnovm is in charge to run [Gno](.) code. It interprets directly the source code and parses it to an AST. After that, it emulates CPU instructions to execute the AST instead of executing it directly. It is able to store the state of a [Realm](./realm.md) and retrieve it on the next execution.