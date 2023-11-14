---
id: 'effective-gno'
---

# Effective Gno

This document provides advice and guidelines for writing effective Gno code.

First, Gno shares many similarities with Go. Therefore, please read ["Effective Go"](https://go.dev/doc/effective_go) first.

## ...

- panic is good
- init() is good
- global variables are good
- NPM-style small and focused libraries are good
- versionning is different
- exporting a variable is unsafe; instead, you should create getters and setters checking for permission to udpate
- safe objects are possible
- exporting an object can be done, but the object needs to be made "securely smart"
- code gen -> is not yet a thing but will be
- unoptimized / gas ineffecient examples
- optimized data structures
- state machines
- patterns to set an initial owner
- TDD
- ship related code to help the review
- write doc for users, not only for developers
- export/unexport things for different reasons
