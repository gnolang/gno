package events

import (
	"chain"

	"gno.land/p/moul/helplink"
)

func Render(_ string) string {
	out := "# Emitting events in Gno\n\n"

	out += "Emitting events in blockchain systems is one of the ways to make off-chain life easier.\n"
	out += "To emit an event, simply use the **Emit()** function found in the `chain` package.\n\n"
	out += "This function takes in string arguments as follows:\n\n"
	out += "```\nchain.Emit(\"EventName\", \"key1\", \"value1\", \"key2\", \"value2\")\n```\n"
	out += "The function takes in a variadic number of key:value pairs after the event name.\n\n"
	out += "Events are stored in block results, and can be listened to via the [tx-indexer](https://github.com/gnolang/tx-indexer).\n\n"

	out += "Click " + helplink.Func("here", "Emit", "customEventValue", "") + " to submit a transaction that will emit an event."

	return out
}

func Emit(cur realm, customEventValue string) {
	chain.Emit("ExampleEvent", "key", customEventValue)
}
