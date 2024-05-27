package subscription

import (
	"fmt"
	"sync"

	"github.com/gnolang/gno/gno.me/event/message"
	"github.com/gnolang/gno/gno.me/state"
	"github.com/google/uuid"
)

type Channel struct {
	sync.RWMutex
	latestSequence string
	subscribers    map[uuid.UUID]*Subscriber
}

func (c *Channel) AddSubscriber(subscriber *Subscriber) {
	c.Lock()
	defer c.Unlock()

	c.subscribers[subscriber.ID()] = subscriber
}

func (c *Channel) RemoveSubscribers(ids []uuid.UUID) {
	c.Lock()
	defer c.Unlock()

	for _, id := range ids {
		delete(c.subscribers, id)
	}
}

func (c *Channel) Broadcast(event *state.Event) ([]uuid.UUID, error) {
	c.RLock()
	defer c.RUnlock()

	var failed []uuid.UUID
	for _, subscriber := range c.subscribers {
		if !subscriber.broadcastTo {
			continue
		}

		msg, err := message.Send{Event: event}.Marshal()
		if err != nil {
			return nil, err
		}

		if err := subscriber.Send(msg); err != nil {
			failed = append(failed, subscriber.ID())
		}
	}

	return failed, nil
}

var (
	channelsLock sync.RWMutex
	channels     map[string]*Channel
)

func init() {
	channels = make(map[string]*Channel)
}

func AddChannel(channel string) {
	channelsLock.Lock()
	defer channelsLock.Unlock()

	fmt.Println("creating subscription channel", channel)
	if _, ok := channels[channel]; !ok {
		channels[channel] = &Channel{
			subscribers: make(map[uuid.UUID]*Subscriber),
		}
	}
}

func GetChannel(channel string) *Channel {
	channelsLock.RLock()
	defer channelsLock.RUnlock()

	return channels[channel]
}
