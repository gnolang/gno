# gnorkle

gnorkle is an attempt at a generalized oracle implementation. The gnorkle `Instance` definitions lives in the `gnorkle` directory. It can be effectively initialized by using the `gnorkle.NewInstance` constructor. The owner of the instance is then able to add feeds with or without whitelists, manage the instance's own whitelist, remove feeds, get the latest feed values, and handle incoming message requests that result in some type of action taking place.

An example of gnorkle in action can be found in the `examples/gno.land/r/gnoland/ghverify` realm. Any realm that wishes to integrate the minimum oracle functionality should include functions that relay incoming messages to the oracle instance's message handler and that can produce the latest feed values.

## Feeds

Feeds that are added to an oracle must implement the `gnorkle.Feed` interface. A feed can be thought of as a mechanism that takes in off-chain data and produces data to be consumed on-chain. The three important components of a feed are:
- Tasks: a feed is composed of one or more `feed.Task` instances. These tasks ultimately instruct an agent what it needs to do to provide data to a feed for ingestion.
- Ingester: a feed should have a `gnorkle.Ingester` instance that handles incoming data from agents. The ingester's role is to perform the correct action on the incoming data, like adding it to an aggregate or storing it for later evaluation.
- Storage: a feed should have a `gnorkle.Storage` instance that manages the state of the data a feed publishes -- data to be consumed on-chain.

A single oracle instance may be composed of many feeds, each of which has a unique ID.

Here is an example of what differences may exist amongst feed implementations:
- Static: a static feed is one that only needs to produce a value once. It ingests values and then publishes the result. Once a single value is published, the state of the feed becomes immutable.
	- Example: a realm wants to integrate football match results for its users. It may embed a static oracle that allows it to publish the match results to the chain. A static feed is a good choice for this because the match results will never change.
- Continuous: a continuous feed can accept and ingest data, continously adding and changing its own internal state based on the data received. It can then publish values on demand based on its current state.
	- Example: a realm wants to provide a verifiable random function. It embeds an oracle that defines tasks and whitelists a select group of trusted agents to provide data that gets combined to produce a random value. A continuous feed is a good choice for this because data may be accepted continously and there is no single static result.
- Periodic:  periodic feed may give all whitelisted agents the opportunity to send data for ingestion within a bounded period of time. After this window closes, the results can be committed and a value is pubished. The process then begins again for the next period.
	- Example: a realm wants to provide weather information to its users. It may choose a group of trusted agents to publish weather data for each half hour interval. This periodic feed can finalize the weather data for each postal code at the end of each interval using the aggregation function defined by the owner of the oracle.

The only feed currently implemented is the `feeds/static.Feed` type.

## Tasks

It's not hard to be a task -- just implement the one method `feed.Task` interface. On-chain task definitions should not do anything other than store data and be able to marshal that data to JSON. Of course, it is also useful if the task marshal's a `type` field as part of the JSON object so an agent is able to know what type of task it is dealing with and what data to expect in the payload.

## Ingesters

An ingester's primary role is to receive data provided by agents and figure out what to do with it. Ingesters must implement the `gnorkle.Ingester` interface. There are currently two message function types that an ingester may want to handle, `message.FuncTypeIngest` and `message.FuncTypeCommit`. The former message type should result in the ingester accumulating data in its own data store, while the latter should use what it has in its data store to publish a feed value to the `gnorkle.Storage` instance provided to it.

The only ingester currently implemented is the `ingesters/single.ValueIngester` type.

## Storage

Storage types are responsible for storing values produced by a feed's ingester. A storage type must implement `gnorkle.Storage`. This type should be able add values to the storage, retrieve the latest value, and retrieve a set of historical values. It is probably a good idea to make the storage bounded.

The only storage currently implemented is the `storage/simple.Storage` type.

## Whitelists

Whitelists are optional but they can be set on both the oracle instance and the feed levels. The absence of a whitelist definition indicates that ALL addresses should be considered to be whitelisted. Otherwise, the presence of a defined whitelist indicates the callers address MUST be in the whitelist in order for the request to succeed. A feed whitelist has precedence over the oracle instance's whitelist. If a feed has a whitelist and the caller is not on it, the call fails. If a feed doesn't have a whitelist but the instance does and the caller is not on it, the call fails. If neither have a whitelist, the call succeeds. The whitlist logic mostly lives in `gnorkle/whitelist.gno` while the only current `gnorkle.Whitelist` implementation is the `agent.Whitelist` type. The whitelist is not owned by the feeds they are associated with in order to not further pollute the `gnorkle.Feed` interface.