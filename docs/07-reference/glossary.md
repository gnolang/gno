# Glossary

<!-- TODO: generate TOC -->

## `p/` - "Pure" packages

A `p/` package denotes a "pure" package within the system. These packages are
crafted as self-contained units of code, capable of being independently imported
and utilized.

Unlike `r/` realms, `p/` packages do not possess states or assets. They are
designed specifically to be called by other packages, whether those packages are
pure or realms.

## `r/` - "Realm" packages

An `r/` realm designates a package endowed with advanced capabilities, referred
to as a "realm".

Realms can accommodate a diverse array of data and functionalities, including
Bank State, Data State, and Address.

They are purposed to furnish and expose features for both user-initiated calls
and as components invoked by other realm packages.
