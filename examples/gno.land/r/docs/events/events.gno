package emitevents

import "std"

func Render(_ string) string {
	out := "# Emitting events in Gno\n\n"

	out += `Emitting events in blockchain systems is one of the ways to make off-chain life easier.
To emit an event, simply use the **Emit()** function found in the **std** package.

This function takes in string arguments as follows:

`
	out += "```\nstd.Emit(\"EventName\", \"key1\", \"value1\", \"key2\", \"value2\")\n```\n"
	out += "The function takes in a variadic number of key:value pairs after the event name.\n\n"
	out += "Events are stored in block results, and can be listened to via the [tx-indexer](https://github.com/gnolang/tx-indexer).\n\n"

	out += "Click [here](/r/docs/events$help&func=Emit&CustomEventValue=) to submit a transaction that will emit an event."

	return out
}

func Emit(CustomEventValue string) {
	std.Emit("ExampleEvent", "key", CustomEventValue)
}
