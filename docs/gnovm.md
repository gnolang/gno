---
state: improvements-needed # The final PoC is working well, but it must evolve performance-wise.
---

# Gnovm

Gnovm is in charge to run [Gno](.) code. 
Firstly, [Gno](.) source code is parsed into an Abstract Syntax Tree (AST).
Then, [Gnovm](./gnovm.md) interprets the AST to execute the code. 
After that, it emulates CPU instructions to execute the AST instead of executing it directly. 
It can store the state of a [Realm](./realm.md) and retrieve it on the next execution.