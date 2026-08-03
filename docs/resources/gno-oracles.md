# Oracles

A realm cannot fetch anything: no HTTP, no files, no external reads. Off-chain
data enters the chain only when someone sends a transaction carrying it. That
someone is an agent: an ordinary off-chain program submitting transactions
from its own address, which the chain sees as just another account. An oracle
is therefore an agreement between a realm and agents it chooses to trust to
send that data.

## The gnorkle framework

[gnorkle](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo/gnorkle)
(`gno.land/p/demo/gnorkle/gnorkle`) structures that agreement so you do not
build the plumbing from scratch. A realm embeds a gnorkle instance and
registers *feeds*, each describing *tasks* for agents to perform. An agent
polls the realm for pending tasks through an entrypoint, does the off-chain
work, and pushes the result back with an ingest message. An *ingester*
validates and commits the value to the feed's storage, and whitelists, per
instance or per feed, control which agents may provide values.

Feeds come in three shapes: a static feed commits one value then locks, a
continuous feed keeps ingesting and publishes on demand, and a periodic feed
aggregates the values received in each time window. Only the static feed is
implemented today.

## Example: publishing a single value

A realm that needs one off-chain fact, say a football match result, embeds an
instance, whitelists an agent, and registers a
[single-value static feed](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo/gnorkle/feeds/static):

```go
import (
	"gno.land/p/demo/gnorkle/feeds/static"
	"gno.land/p/demo/gnorkle/gnorkle"
)

var oracle *gnorkle.Instance

func init() {
	oracle = gnorkle.NewInstance()
	// "" targets the instance-level whitelist.
	oracle.AddToWhitelist("", []string{"g1trustedagent..."})
	oracle.AddFeeds(static.NewSingleValueFeed("match-1", "string", matchTask{}))
}

// The task carries the data the agent needs to do the off-chain work.
type matchTask struct{}

func (matchTask) MarshalJSON() ([]byte, error) {
	return []byte(`{"type":"match_result","match":"match-1"}`), nil
}

func GnorkleEntrypoint(cur realm, msg string) string {
	result, err := oracle.HandleMessage(msg, nil)
	if err != nil {
		panic(err)
	}
	return result
}

func Result() string {
	value, _, consumable, err := oracle.GetFeedValue("match-1")
	if err != nil || !consumable {
		return "pending"
	}
	return value.String
}
```

The agent drives the flow with two messages to `GnorkleEntrypoint`.
`request` returns the JSON definitions of every feed the agent is whitelisted
for, tasks included. `ingest,match-1,2-1` submits the value: a single-value
feed commits it on first ingestion and locks, so the result can never be
overwritten. `Result` then serves the committed value to any caller.

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
