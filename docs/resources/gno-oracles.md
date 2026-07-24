# Oracles

A realm cannot fetch anything: no HTTP, no files, no external reads. Off-chain
data enters the chain only when someone sends a transaction carrying it. An
oracle is therefore an agreement between a realm and off-chain agents it
chooses to trust to send that data.

## The gnorkle framework

[gnorkle](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo/gnorkle)
(`gno.land/p/demo/gnorkle/gnorkle`) structures that agreement so you do not
build the plumbing from scratch. A realm embeds a gnorkle instance and
registers *feeds*, each describing *tasks* for agents to perform. An agent
polls the realm for pending tasks through an entrypoint, does the off-chain
work, and pushes the result back with an ingest message. An *ingester*
validates and commits the value to the feed's storage, and whitelists, per
instance or per feed, control which agents may provide values.

## Example: verifying a GitHub identity

[ghverify](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/gnoland/ghverify)
(`gno.land/r/gnoland/ghverify`) is a complete deployed oracle. A user calls
`RequestVerification` with their GitHub handle, which registers a feed. An
off-chain agent picks up the task, checks that the handle controls a
repository containing the user's address, and ingests the result. The realm
then serves the verified handle-to-address mapping to any caller.

## Trust model

The chain never verifies the off-chain fact, only that a whitelisted agent
attested to it. The whitelist is the trust root: whoever controls the agents
controls the data. There is no built-in price feed, so a realm that moves
funds based on a fed value is only as secure as whoever provides that value.
See [Effective Gno](./effective-gno.md#bring-off-chain-data-on-chain-with-oracles)
for how to design around this.
