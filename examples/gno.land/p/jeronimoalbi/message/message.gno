package message

// TopicAll defines a topic for all types of message.
// This topic can be used to subscribe to message for all topics.
const TopicAll Topic = "*"

type (
	// Topic defines a type for message topics.
	Topic string

	// Callback defines a type for message callbacks.
	Callback func(Message)

	// Message defines a type for published messages.
	Message struct {
		// Topic is the message topic.
		Topic Topic

		// Data contains optional message data.
		Data any
	}

	// Publisher defines an interface for message publishers.
	Publisher interface {
		// Publish publishes a message for a topic.
		Publish(_ Topic, data any) error
	}

	// Subscriber defines an interface for message subscribers.
	Subscriber interface {
		// Subscribe subscribes to messages published for a topic.
		// It returns the callback ID within the topic.
		Subscribe(Topic, Callback) (id int, _ error)

		// Unsubscribe unsubscribes a callback from a message topic.
		// ID is the callback ID within the topic, returned on subscription.
		Unsubscribe(_ Topic, id int) (unsubscribed bool, _ error)
	}
)
