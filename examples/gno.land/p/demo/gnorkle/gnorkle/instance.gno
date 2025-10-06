package gnorkle

import (
	"chain/runtime"
	"errors"
	"strings"

	"gno.land/p/demo/gnorkle/agent"
	"gno.land/p/demo/gnorkle/feed"
	"gno.land/p/demo/gnorkle/message"
	"gno.land/p/nt/avl"
)

// Instance is a single instance of an oracle.
type Instance struct {
	feeds     *avl.Tree
	whitelist agent.Whitelist
}

// NewInstance creates a new instance of an oracle.
func NewInstance() *Instance {
	return &Instance{
		feeds: avl.NewTree(),
	}
}

func assertValidID(id string) error {
	if len(id) == 0 {
		return errors.New("feed ids cannot be empty")
	}

	if strings.Contains(id, ",") {
		return errors.New("feed ids cannot contain commas")
	}

	return nil
}

func (i *Instance) assertFeedDoesNotExist(id string) error {
	if i.feeds.Has(id) {
		return errors.New("feed already exists")
	}

	return nil
}

// AddFeeds adds feeds to the instance with empty whitelists.
func (i *Instance) AddFeeds(feeds ...Feed) error {
	for _, feed := range feeds {
		if err := assertValidID(feed.ID()); err != nil {
			return err
		}

		if err := i.assertFeedDoesNotExist(feed.ID()); err != nil {
			return err
		}

		i.feeds.Set(
			feed.ID(),
			FeedWithWhitelist{
				Whitelist: new(agent.Whitelist),
				Feed:      feed,
			},
		)
	}

	return nil
}

// AddFeedsWithWhitelists adds feeds to the instance with the given whitelists.
func (i *Instance) AddFeedsWithWhitelists(feeds ...FeedWithWhitelist) error {
	for _, feed := range feeds {
		if err := i.assertFeedDoesNotExist(feed.ID()); err != nil {
			return err
		}
		if err := assertValidID(feed.ID()); err != nil {
			return err
		}

		i.feeds.Set(
			feed.ID(),
			FeedWithWhitelist{
				Whitelist: feed.Whitelist,
				Feed:      feed,
			},
		)
	}

	return nil
}

// RemoveFeed removes a feed from the instance.
func (i *Instance) RemoveFeed(id string) {
	i.feeds.Remove(id)
}

// PostMessageHandler is a type that allows for post-processing of feed state after a feed
// ingests a message from an agent.
type PostMessageHandler interface {
	Handle(i *Instance, funcType message.FuncType, feed Feed) error
}

// HandleMessage handles a message from an agent and routes to either the logic that returns
// feed definitions or the logic that allows a feed to ingest a message.
//
// TODO: Consider further message types that could allow administrative action such as modifying
// a feed's whitelist without the owner of this oracle having to maintain a reference to it.
func (i *Instance) HandleMessage(msg string, postHandler PostMessageHandler) (string, error) {
	caller := string(runtime.OriginCaller())

	funcType, msg := message.ParseFunc(msg)

	switch funcType {
	case message.FuncTypeRequest:
		return i.GetFeedDefinitions(caller)

	default:
		id, msg := message.ParseID(msg)
		if err := assertValidID(id); err != nil {
			return "", err
		}

		feedWithWhitelist, err := i.getFeedWithWhitelist(id)
		if err != nil {
			return "", err
		}

		if !addressIsWhitelisted(&i.whitelist, feedWithWhitelist, caller, nil) {
			return "", errors.New("caller not whitelisted")
		}

		if err := feedWithWhitelist.Ingest(funcType, msg, caller); err != nil {
			return "", err
		}

		if postHandler != nil {
			postHandler.Handle(i, funcType, feedWithWhitelist)
		}
	}

	return "", nil
}

func (i *Instance) getFeed(id string) (Feed, error) {
	untypedFeed, ok := i.feeds.Get(id)
	if !ok {
		return nil, errors.New("invalid ingest id: " + id)
	}

	feed, ok := untypedFeed.(Feed)
	if !ok {
		return nil, errors.New("invalid feed type")
	}

	return feed, nil
}

func (i *Instance) getFeedWithWhitelist(id string) (FeedWithWhitelist, error) {
	untypedFeedWithWhitelist, ok := i.feeds.Get(id)
	if !ok {
		return FeedWithWhitelist{}, errors.New("invalid ingest id: " + id)
	}

	feedWithWhitelist, ok := untypedFeedWithWhitelist.(FeedWithWhitelist)
	if !ok {
		return FeedWithWhitelist{}, errors.New("invalid feed with whitelist type")
	}

	return feedWithWhitelist, nil
}

// GetFeedValue returns the most recently published value of a feed along with a string
// representation of the value's type and boolean indicating whether the value is
// okay for consumption.
func (i *Instance) GetFeedValue(id string) (feed.Value, string, bool, error) {
	foundFeed, err := i.getFeed(id)
	if err != nil {
		return feed.Value{}, "", false, err
	}

	value, valueType, consumable := foundFeed.Value()
	return value, valueType, consumable, nil
}

// GetFeedDefinitions returns a JSON string representing the feed definitions for which the given
// agent address is whitelisted to provide values for ingestion.
func (i *Instance) GetFeedDefinitions(forAddress string) (string, error) {
	instanceHasAddressWhitelisted := !i.whitelist.HasDefinition() || i.whitelist.HasAddress(forAddress)

	buf := new(strings.Builder)
	buf.WriteString("[")
	first := true
	var err error

	// The boolean value returned by this callback function indicates whether to stop iterating.
	i.feeds.Iterate("", "", func(_ string, value any) bool {
		feedWithWhitelist, ok := value.(FeedWithWhitelist)
		if !ok {
			err = errors.New("invalid feed type")
			return true
		}

		// Don't give agents the ability to try to publish to inactive feeds.
		if !feedWithWhitelist.IsActive() {
			return false
		}

		// Skip feeds the address is not whitelisted for.
		if !addressIsWhitelisted(&i.whitelist, feedWithWhitelist, forAddress, &instanceHasAddressWhitelisted) {
			return false
		}

		var taskBytes []byte
		if taskBytes, err = feedWithWhitelist.Feed.MarshalJSON(); err != nil {
			return true
		}

		// Guard against any tasks that shouldn't be returned; maybe they are not active because they have
		// already been completed.
		if len(taskBytes) == 0 {
			return false
		}

		if !first {
			buf.WriteString(",")
		}

		first = false
		buf.Write(taskBytes)
		return false
	})

	if err != nil {
		return "", err
	}

	buf.WriteString("]")
	return buf.String(), nil
}
