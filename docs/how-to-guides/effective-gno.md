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
- exported global variables are unsafe
- safe objects are possible

---

- different goal for exported variables and methods

## Insecure Usages

- exporting a variable; instead, you should create getters and setters checking for permission to udpate

...

## Inefficient Usages

unoptimized / gas ineffecient

...

## Good Gno Patterns

...

## Bad Gno Patterns

inefficient, dangerous.
