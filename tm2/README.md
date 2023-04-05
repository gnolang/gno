# Tendermint2

**Disclaimer: Tendermint2 is currently part of the Gno monorepo for streamlined development. Once Gno.land is on the mainnet, Tendermint2 will operate independently, including for governance, on https://github.com/tendermint/tendermint2.**

## Mission

 * make awesome software with modular components.
 * crypto p2p swiss armyknife for human liberation.

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
 * Modular dependencies - whereever reasonable, make components modular.
 * Completeness - software projects that don't become finished are projects
   that are forever vulnerable. One of the primary goals of the Gno language
   and related works is to become finished within a reasonable timeframe.

## What is already proposed for Tendermint2:

* Complete Amino. -> multiplier of productivity for SDK development, to not
  have to think about protobuf at all. Use "genproto" to even auto-generate
  proto3 for encoding/decoding optimization through protoc. // MISSION: be the
  basis for improving the encoding standard from proto3, because proto3
  length-prefixing is slow, and we need "proto4" or "amino2". // LOOK at the
  auto-generated proto files!
  https://github.com/gnolang/gno/blob/master/pkgs/bft/consensus/types/cstypes.proto
  for example. // There was work to remove this from the CosmosSDK because
  Amino wasn't ready, but now that it is, it makes sense to incorporate it into
  Tendermint2.

* Remove EvidenceReactor, Evidence, Violation -> we need to make it easy to
  create alt mempool reactors. We "kill two birds with one stone" by
  implementing evidence as a first-class mempool lane. The authors of "ABCI++"
  have a different set of problems to solve, so we should do both! Tendermint++
  and Tendermint2.

* Fix address size to 20 bytes -> 160 is sufficient, and fixing it brings
  optimizations.

* General versionset system for handshake negotiation. -> So Tendermint2 can be
  used as basis for other p2p applications.

* EventBus -> EventSwitch. -> For indexing, use an external system. This keeps
  Tendermint2 minimal, allowing integration with plugin modules, without having
  any internal implementation at all. EventSwitch is also simpler, and
  synchronous, and this keeps the Tendermint tests deterministic. There is no
  performance need to do anything else than keep the Tendermint protocol
  synchronous. (If there is, because of massive validator numbers for whatever
  reason, then it should be a fork of Tendermint with a unique & distinct name,
  and would be under the same taxonomy of Tendermint).

* Fix nondeterminism in consensus tests -> in relation to the above.

* Add "MaxDataBytes" for total tx data size limitation. -> The previous way of
  limiting the total block size may result in unexpected behavior with changes
  in validator size. We should err to allocate room for each module seperately,
  to ensure availability.

* Remove external dependencies like prometheus. -> Any metrics and events
  should be plugged in through the implementation of interfaces. This may
  involve picking out the client logic from prometheus, but even if so it would
  be forked into Tendermint2 and be audited like anything else.

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
