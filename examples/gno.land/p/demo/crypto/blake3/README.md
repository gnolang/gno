blake3
------

This is a `gno` port of [lukechampine.com/blake3](https://lukechampine.com/blake3) [version 1.3.0](https://github.com/lukechampine/blake3/releases/tag/v1.3.0)

Notable changes to this implementation:

- Inlining testing, since we don't have disk access
- removal of CPU optimizations (go native only)
- added type casting on untyped `const` where needed

Goals:
- Keep this as true to implementation as possible
- Keep tests as similar as possible
- Provide a performant and correct implementation of blake3
