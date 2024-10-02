---
id: tendermint2
---

# Tendermint2

**Disclaimer: Tendermint2 is currently part of the Gno monorepo for streamlined development.**

**Once gno.land is on the mainnet, Tendermint2 will operate independently, including for governance,
on https://github.com/tendermint/tendermint2.**

## Problems

* Open source is open for subversion.
* Incentives and mission are misaligned.
* Need directory & forum for Tendermint/SDK forks.

## Partial Solution: adopt principles

* Simplicity of design.
* The code is the spec.
* Minimal code - keep total footprint small.
* Minimal dependencies - all dependencies must get audited, and become part of
  the repo.
* Modular dependencies - wherever reasonable, make components modular.
* Completeness - software projects that don't become finished are projects
  that are forever vulnerable. One of the primary goals of the Gno language
  and related works is to become finished within a reasonable timeframe.

## What is already proposed for Tendermint2:

* Complete Amino. -> multiplier of productivity for SDK development, to not
  have to think about protobuf at all. Use "genproto" to even auto-generate
  proto3 for encoding/decoding optimization through protoc.
    - MISSION: be the basis for improving the encoding standard from proto3, because
      proto3 length-prefixing is slow, and we need "proto4" or "amino2".
    - LOOK at the [auto-generated proto files](https://github.com/gnolang/gno/blob/master/tm2/pkg/bft/consensus/consensus.proto)!
    - There was work to remove this from the CosmosSDK because
      Amino wasn't ready, but now that it is, it makes sense to incorporate it into
      Tendermint2.


* Remove EvidenceReactor, Evidence, Violation:

  We need to make it easy to create alt mempool reactors.

  We "kill two birds with one stone" by implementing evidence as a first-class mempool lane.

  The authors of "ABCI++" have a different set of problems to solve, so we should do both! Tendermint++
  and Tendermint2.


* Fix address size to 20 bytes -> 160 is sufficient, and fixing it brings optimizations.


* General versionset system for handshake negotiation. -> So Tendermint2 can be
  used as basis for other P2P applications.


* EventBus -> EventSwitch. -> For indexing, use an external system.

  To ensure Tendermint2 remains minimal and easily integrated with plugin modules, there is no internal implementation.

  The use of an EventSwitch makes the process simpler and synchronous, which maintains the determinism of Tendermint
  tests.

  Keeping the Tendermint protocol synchronous is sufficient for optimal performance.

  However, if there is a need for asynchronous processing due to an exceptionally large number of validators, it should
  be a separate fork with a unique name under the same taxonomy as Tendermint.


* Fix nondeterminism in consensus tests -> in relation to the above.

* Add "MaxDataBytes" for total tx data size limitation.

  To avoid unexpected behavior caused by changes in validator size, it's best to allocate room for each module
  separately instead of limiting the total block size as we did before.

This way, we can ensure that there's enough space for all modules.

* Remove external dependencies like prometheus
  To ensure accuracy, all metrics and events should be integrated through interfaces. This may require extracting client
  logic from Prometheus, but it will be incorporated into Tendermint2 and undergo the same auditing process as
  everything else.

* General consensus/WAL -> a WAL is useful enough to warrant being a re-usable
  module.

* Remove GRPC -> GRPC support should be plugged in (say in a GRPC fork of
  Tendermint2), so alternative RPC protocols can likewise be. Tendermint2 aims
  to be independent of the Protobuf stack so that it can retain freedom for
  improving its codec.

* Remove dependency on viper/cobra -> I have tried to strip out what we don't
  use of viper/cobra for minimalism, but could not; and viper/cobra is one
  prime target for malware to be introduced. Rather than audit viper/cobra,
  Tendermint2 implements a cli convention for Go-structure-based flags and cli;
  so if you still want to use viper/cobra you can do so by translating flags to
  an options struct.

* Question: Which projects use ABCI sockets besides CosmosSDK derived chains?

## Roadmap

First, we create a multi-organizational team for Tendermint2 &
TendermintCore/++ development. We will maintain a fork of the Tendermint++ repo
and suggest changes upstream based on our work on Tendermint2, while also
porting necessary fixes from Tendermint++ over to Tendermint2.

We will also reach out to ecosystem partners and survey and create a
directory/taxonomy for Tendermint and CosmosSDK derivatives and manage a forum
for interfork collaboration.

Ideally, Tendermint2 and TendermintCore/++ merge into one.

## Challenge

Either make a PR to Gaia/CosmosSDK/TendermintCore to be like Tendermint2, or
make a PR to Tendermint2 to import a feature or fix of TendermintCore.
