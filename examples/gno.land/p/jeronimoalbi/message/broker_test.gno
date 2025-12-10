package message_test

import (
	"testing"

	"gno.land/p/jeronimoalbi/expect"
	"gno.land/p/jeronimoalbi/message"
)

var (
	_ message.Subscriber = (*message.Broker)(nil)
	_ message.Publisher  = (*message.Broker)(nil)
)

func TestBrokerTopics(t *testing.T) {
	broker := message.NewBroker()
	expect.
		Value(t, len(broker.Topics())).
		AsInt().
		ToEqual(0)

	cb := func(message.Message) {}
	broker.Subscribe("foo", cb)
	broker.Subscribe("bar", cb)
	broker.Subscribe("baz", cb)
	broker.Subscribe(message.TopicAll, cb)
	topics := broker.Topics()

	expect.
		Value(t, len(topics)).
		AsInt().
		ToEqual(3)
	expect.
		Value(t, string(topics[0])).
		AsString().
		ToEqual("bar")
	expect.
		Value(t, string(topics[1])).
		AsString().
		ToEqual("baz")
	expect.
		Value(t, string(topics[2])).
		AsString().
		ToEqual("foo")
}

func TestBrokerPublish(t *testing.T) {
	tests := []struct {
		name               string
		subscribe, publish message.Topic
		data               any
		message            *message.Message
		err                error
	}{
		{
			name:      "publishes subscribed topic",
			subscribe: "foo",
			publish:   "foo",
			data:      "foo's data",
			message: &message.Message{
				Topic: "foo",
				Data:  "foo's data",
			},
		},
		{
			name:      "publishes all topics",
			subscribe: message.TopicAll,
			publish:   "foo",
			data:      "foo's data",
			message: &message.Message{
				Topic: "foo",
				Data:  "foo's data",
			},
		},
		{
			name:      "invalid topic",
			subscribe: "foo",
			publish:   message.TopicAll,
			err:       message.ErrInvalidTopic,
		},
		{
			name:      "no topic",
			subscribe: "foo",
			publish:   "",
			err:       message.ErrRequiredTopic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var msg *message.Message
			broker := message.NewBroker()
			broker.Subscribe(tt.subscribe, func(m message.Message) { msg = &m })

			// Act
			err := broker.Publish(tt.publish, tt.data)

			// Assert
			if tt.err != nil {
				expect.
					Func(t, func() error { return err }).
					WithFailPrefix("expect a publish error").
					ToFail().
					WithError(tt.err)
				expect.
					Value(t, istypednil(msg)).
					WithFailPrefix("expect callback not to be called").
					AsBoolean().
					ToBeTruthy()
				return
			}

			expect.
				Value(t, err).
				WithFailPrefix("expect no publish error").
				ToBeNil()
			expect.
				Value(t, msg).
				WithFailPrefix("expect callback to be called").
				Not().ToBeNil()
			expect.
				Value(t, string(msg.Topic)).
				WithFailPrefix("expect message topic to match").
				AsString().
				ToEqual(string(tt.message.Topic))
			expect.
				Value(t, msg.Data).
				WithFailPrefix("expect message data to match").
				AsString().
				ToEqual(tt.message.Data.(string))
		})
	}
}

func TestBrokerSubscribe(t *testing.T) {
	cb := func(message.Message) {}
	tests := []struct {
		name     string
		setup    func(*message.Broker)
		topic    message.Topic
		id       int64
		callback message.Callback
		err      error
	}{
		{
			name:     "single subscription",
			topic:    "foo",
			id:       1,
			callback: cb,
		},
		{
			name: "existing subscriptions",
			setup: func(b *message.Broker) {
				b.Subscribe("foo", cb)
			},
			topic:    "foo",
			id:       2,
			callback: cb,
		},
		{
			name: "other subscription topics",
			setup: func(b *message.Broker) {
				b.Subscribe("foo", cb)
			},
			topic:    "bar",
			id:       1,
			callback: cb,
		},
		{
			name:  "no topic",
			topic: "",
			err:   message.ErrRequiredTopic,
		},
		{
			name:     "no callback",
			topic:    "foo",
			callback: nil,
			err:      message.ErrRequiredCallback,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			broker := message.NewBroker()
			if tt.setup != nil {
				tt.setup(broker)
			}

			// Act
			id, err := broker.Subscribe(tt.topic, tt.callback)

			// Assert
			if tt.err != nil {
				expect.
					Func(t, func() error { return err }).
					WithFailPrefix("expect a subscribe error").
					ToFail().
					WithError(tt.err)
				expect.
					Value(t, id).
					WithFailPrefix("expect zero ID").
					AsInt().
					ToEqual(0)
				return
			}

			expect.
				Value(t, err).
				WithFailPrefix("expect no subscribe error").
				ToBeNil()
			expect.
				Value(t, id).
				WithFailPrefix("expect ID to match").
				AsInt().
				ToEqual(tt.id)
		})
	}
}

func TestBrokerUnsubscribe(t *testing.T) {
	cb := func(message.Message) {}
	tests := []struct {
		name   string
		setup  func(*message.Broker)
		topic  message.Topic
		id     int
		errMsg string
	}{
		{
			name: "single subscription",
			setup: func(b *message.Broker) {
				b.Subscribe("foo", cb)
			},
			topic: "foo",
			id:    1,
		},
		{
			name: "multiple subscriptions",
			setup: func(b *message.Broker) {
				b.Subscribe("foo", cb)
				b.Subscribe("foo", cb)
			},
			topic: "foo",
			id:    2,
		},
		{
			name: "other subscription topics",
			setup: func(b *message.Broker) {
				b.Subscribe("foo", cb)
				b.Subscribe("bar", cb)
			},
			topic: "foo",
			id:    1,
		},
		{
			name:   "not found",
			topic:  "foo",
			id:     1,
			errMsg: "message topic not found: foo",
		},
		{
			name:   "no topic",
			topic:  "",
			errMsg: message.ErrRequiredTopic.Error(),
		},
		{
			name:   "no subscription ID",
			topic:  "foo",
			id:     0,
			errMsg: message.ErrRequiredSubscriptionID.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			broker := message.NewBroker()
			if tt.setup != nil {
				tt.setup(broker)
			}

			// Act
			unsubscribed, err := broker.Unsubscribe(tt.topic, tt.id)

			// Assert
			if tt.errMsg != "" {
				expect.
					Func(t, func() error { return err }).
					WithFailPrefix("expect a subscribe error").
					ToFail().
					WithMessage(tt.errMsg)
				expect.
					Value(t, unsubscribed).
					WithFailPrefix("expect unsubscribe to fail").
					AsBoolean().
					ToBeFalsy()
				return
			}

			expect.
				Value(t, err).
				WithFailPrefix("expect no unsubscribe error").
				ToBeNil()
			expect.
				Value(t, unsubscribed).
				WithFailPrefix("expect unsubscribe to succeed").
				AsBoolean().
				ToBeTruthy()
		})
	}
}
