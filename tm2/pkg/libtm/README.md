## Overview

`libtm` is a simple, minimal and compact Go library that implements the Tendermint consensus engine.

The implementation is based on Algorithm 1, of
the [Tendermint consensus whitepaper](https://arxiv.org/pdf/1807.04938.pdf) and
(more broadly) the [official Tendermint wiki](https://github.com/tendermint/tendermint/wiki).

There are some implementation design decisions taken by the package authors:

- it doesn't manage validator sets internally
- it doesn't implement a networking layer, or any kind of broadcast communication
- it doesn't assume, or implement, any kind of signature manipulation logic

All of these responsibilities are left to the calling context, in the form of interface implementations.
The reason for these choices is simple - _to keep the library minimal_.

> [!NOTE]
> We aim to make [libtm](https://github.com/gnolang/libtm) an independent project, both in terms of repository and
> governance. This will be pursued once we have successfully integrated `libtm` with `tm2` and demonstrated that the
> codebase and its API are stable and reliable for broader use. Until this integration is complete and stability is
> confirmed, `libtm` will continue to be improved upon within the current structure, in the gno monorepo.

### What this library is

This library is meant to be used as a consensus engine base for any distributed system that needs such functionality.
It is not exclusively made for the blockchain context -- you will find no mention or assumptions of blockchains in the
source code.

### What this library is not

This library is _not_ meant to replace your entire consensus setup.

As mentioned before, certain design decisions have been taken to keep the source code minimal, which results in the
calling context being a bit more involved in orchestration.

Please, before deciding to utilize this library in your project, understand the different moving parts and their
requirements.

## Installation

To get up and running with the `libtm` package, you can add it to your project using:

```shell
go get -u github.com/gnolang/libtm
```

Currently, the minimum required go version is `go 1.22`.

## Usage Examples

```go
package main

import (
	"context"

	"github.com/gnolang/libtm/core"
	"github.com/gnolang/libtm/messages/types"
)

// Verifier implements the libtm Verifier interface.
// Verifier is an abstraction over the outer consensus calling context
// that has access to validator set information
type Verifier struct {
	// ...
}

// Node implements the libtm Node interface.
// The Node interface is an abstraction over a single entity (current process) that runs
// the consensus algorithm
type Node struct {
	// ...
}

// Broadcast implements the libtm Broadcast interface.
// Broadcast is an abstraction over the networking / message sharing interface
// that enables message passing between validators
type Broadcast struct {
	// ...
}

// Signer implements the libtm Signer interface.
// Signer is an abstraction over the signature manipulation process
type Signer struct {
	// ...
}

// ...

func main() {
	// verifier, node, broadcast, signer, opts
	var (
		verifier  = NewVerifier()
		node      = NewNode()
		broadcast = NewBroadcast()
		signer    = NewSigner()
	)

	tm := core.NewTendermint(
		verifier,
		node,
		broadcast,
		signer,
	)

	height := uint64(1)
	ctx, cancelFn := context.WithCancel(context.Background())

	go func() {
		// Run the consensus sequence for the given height.
		// When the method returns the finalized proposal, that means that
		// consensus was reached within the given height (in any round)
		finalizedProposal := tm.RunSequence(ctx, height)

		// Use the finalized proposal
		// ...
	}()

	go func() {
		// Pipe messages into the consensus engine
		var proposalMessage *types.ProposalMessage

		if err := tm.AddProposalMessage(proposalMessage); err != nil {
			// ...
		}

		// ...

		var prevoteMessage *types.PrevoteMessage

		if err := tm.AddPrevoteMessage(prevoteMessage); err != nil {
			// ...
		}

		// ...

		var precommitMessage *types.PrecommitMessage

		if err := tm.AddPrecommitMessage(precommitMessage); err != nil {
			// ...
		}
	}()

	// ...

	// Stop the sequence at any time by cancelling the context
	cancelFn()
}

```

### Additional Options

You can utilize additional options when creating the `Tendermint` consensus engine instance:

- `WithLogger` - specifies the logger for the Tendermint consensus engine (slog)
- `WithProposeTimeout` specifies the propose state timeout
- `WithPrevoteTimeout` specifies the prevote state timeout
- `WithPrecommitTimeout` specifies the precommit state timeout
