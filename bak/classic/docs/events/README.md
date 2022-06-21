Tendermint/Classic employs a simple synchronous EventSwitch for publishing
event messages.  This event pubsub system is distinct from the asynchronous and
more complex PubSub/EventBus system currently in Tendermint/Core, and is one of
the key differentiating factors (as of yet) between the two projects.

Maintaining a synchronous event system makes testing of the consensus engine
more robust, as it requires the test to specify *all* the events transmitted
*in order* -- for otherwise the consensus state machine would halt.

As it does for tests, *the synchronous event model affords event listeners with
deterministic behavior*.  Most event listeners *require* that all relevant
events are received in order.  If the event listener connection's buffers are
full, it is better to close the connection to signify an error, than to skip
some events and resume silently.

This works for a limited number of non-blocking event listeners, but not all
event listeners are non-blocking.  All web-based listeners (e.g. websocket
subscribers) and some internal listeners with disk IO or expensive computation
require an asynchronous pubsub system.  In Tendermint/Classic, such
asynchronous pubsub services are listeners of the underlying synchronous
EventSwitch system.

For websocket event subscribers we also want to provide idempotency.  Listeners
should be able to reconnect and resume streaming events since the last event
received.  The best way to ensure idempotency in this way is to implement a
fifo cache in memory or in the filesystem of all the recent events, and for
each event listener's goroutine to source events from this buffer.  This is
another divergence from Tendermint/Core's pubsub design.  In accordance with
the principle of modularity, *the architecture should be such that an external
process can receive all the events via a websocket stream, to be filtered and
broadcast to all of its subscribing listeners*.

TODO: insert documenttion on this after it is written.

For filtering events, instead of employing the custom query language as in
Tendermint/Core, we will employ a limited subset of Go expression langauges.
The event object is also much simplified as compared to Tendermint/Classic's
event+object+attributes model, and leverages go-amino-x to allow arbitrary Go
structures to serve as an event.

TODO: insert documentation on this after it is written.
