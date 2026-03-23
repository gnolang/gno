package main

import (
	"gno.land/p/jeronimoalbi/message"
)

func main() {
	// Create a message broker and a generic message callback
	broker := message.NewBroker()
	cb := func(m message.Message) {
		println("topic triggered: " + string(m.Topic))
		println("topic data: " + m.Data.(string))
	}

	// Subscribe to a couple of events
	_, err := broker.Subscribe("eventA", cb)
	if err != nil {
		panic(err)
	}

	_, err = broker.Subscribe("eventB", cb)
	if err != nil {
		panic(err)
	}

	// Subscribe to an event and then unsubscribe from it
	id, err := broker.Subscribe("eventC", cb)
	if err != nil {
		panic(err)
	}

	_, err = broker.Unsubscribe("eventC", id)
	if err != nil {
		panic(err)
	}

	// Subscribe to all events
	_, err = broker.Subscribe(message.TopicAll, func(m message.Message) {
		println("catchall topic triggered: " + string(m.Topic))
	})
	if err != nil {
		panic(err)
	}

	// List broker topics
	println("topics:")
	for _, topic := range broker.Topics() {
		println("- " + string(topic))
	}

	// Publish events
	println()
	if err = broker.Publish("eventA", "A"); err != nil {
		panic(err)
	}

	println()
	if err = broker.Publish("eventB", "B"); err != nil {
		panic(err)
	}
}

// Output:
// topics:
// - eventA
// - eventB
// - eventC
//
// topic triggered: eventA
// topic data: A
// catchall topic triggered: eventA
//
// topic triggered: eventB
// topic data: B
// catchall topic triggered: eventB
