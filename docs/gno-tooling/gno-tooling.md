---
id: gno-tooling
---

# Gno Tooling

Welcome to the **Gno Tooling** section for Gno. This section outlines programs & tools that are commonly used when developing applications with Gno.

## Gno Command Syntax Guide

### gno [subcommand] [flags] [arg...]

#### Subcommand

The gno command consists of various purpose-built subcommands.

- `gno {mod}` : manage gno.mod
- `gno {mod} {download} : download modules to local cache

#### Flags

Options of the subcommand.

- `gno mod download [-remote]` : remote for fetching gno modules

#### Arg

The actual value of the flag .

- `gno mod download -remote {rpc.gno.land:26657}`
