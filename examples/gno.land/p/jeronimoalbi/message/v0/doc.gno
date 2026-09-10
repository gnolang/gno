// Package message provides a simple message broker implementation.
//
// The message broker is a Pub/Sub one. It implements two different interfaces,
// `Publisher` and `Subscriber`, which are also defined within this package.
//
// Published messages contain the topic where they are published and optional
// message data.
//
// Subscribe to an event:
//
//	broker := message.NewBroker()
//	subID, err := broker.Subscribe("EventName", func(msg message.Message) {
//	   println("EventName has been triggered")
//	   println(msg.Data)
//	})
//	if err != nil {
//	   panic(err)
//	}
//
// Unsubscribe from an event:
//
//	unsubscribed, err := broker.Unsubscribe("EventName", subID)
//	if err != nil {
//	   panic(err)
//	}
//
//	if !unsubscribed {
//	   panic("subscription not found")
//	}
//
// Publish an event:
//
//	err := broker.Publish("EventName", "Example event data")
//	if err != nil {
//	   panic(err)
//	}
package message
