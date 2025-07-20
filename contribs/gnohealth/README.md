# `gnohealth`: Gno Health Check CLI

**`gnohealth`** is a small command-line utility that provides health checks and diagnostic commands to verify that key components of the Gno toolchain and runtime are functioning as expected.

## Features

- Provides health-related subcommands (starting with `timestamp`)
- Designed for CI systems, developers, and system diagnostics
- Can be extended with additional testable subcommands over time

## Usage

### Run the default health check (timestamp):

````
gnohealth timestamp
```

This verifies basic system behavior and prints the current timestamp as reported by the Gno runtime environment.

## Installation

```
go install github.com/gnolang/gno/contribs/gnohealth@latest
```

## Subcommands

- `timestamp` – Prints a timestamp, serving as a basic health check or ping mechanism

