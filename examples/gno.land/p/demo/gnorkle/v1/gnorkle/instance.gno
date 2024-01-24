package gnorkle

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/gnorkle/v1/agent"
	"gno.land/p/demo/gnorkle/v1/feed"
	"gno.land/p/demo/gnorkle/v1/message"
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

func assertValidID(id string) {
	if len(id) == 0 {
		panic("feed ids cannot be empty")
	}

	if strings.Contains(id, ",") {
		panic("feed ids cannot contain commas")
	}
}

// AddFeeds adds feeds to the instance with empty whitelists.
func (i *Instance) AddFeeds(feeds ...Feed) {
	for _, feed := range feeds {
		assertValidID(feed.ID())
		i.feeds.Set(
			feed.ID(),
			FeedWithWhitelist{
				Whitelist: new(agent.Whitelist),
				Feed:      feed,
			},
		)
	}
}

// AddFeedsWithWhitelists adds feeds to the instance with the given whitelists.
func (i *Instance) AddFeedsWithWhitelists(feeds ...FeedWithWhitelist) {
	for _, feed := range feeds {
		assertValidID(feed.ID())
		i.feeds.Set(
			feed.ID(),
			FeedWithWhitelist{
				Whitelist: feed.Whitelist,
				Feed:      feed,
			},
		)
	}
}

// RemoveFeed removes a feed from the instance.
func (i *Instance) RemoveFeed(id string) {
	i.feeds.Remove(id)
}

// PostMessageHandler is a type that allows for post-processing of feed state after a feed
// ingests a message from an agent.
type PostMessageHandler interface {
	Handle(i *Instance, funcType message.FuncType, feed Feed)
}

// HandleMessage handles a message from an agent and routes to either the logic that returns
// feed definitions or the logic that allows a feed to ingest a message.
//
// TODO: Consider further message types that could allow administrative action such as modifying
// a feed's whitelist without the owner of this oracle having to maintain a reference to it.
func (i *Instance) HandleMessage(msg string, postHandler PostMessageHandler) string {
	caller := string(std.GetOrigCaller())

	funcType, msg := message.ParseFunc(msg)

	switch funcType {
	case message.FuncTypeRequest:
		return i.GetFeedDefinitions(caller)

	default:
		id, msg := message.ParseID(msg)
		assertValidID(id)

		feedWithWhitelist := i.getFeedWithWhitelist(id)
		if !addressIsWhitelisted(&i.whitelist, feedWithWhitelist, caller, nil) {
			panic("caller not whitelisted")
		}

		feedWithWhitelist.Ingest(funcType, msg, caller)

		if postHandler != nil {
			postHandler.Handle(i, funcType, feedWithWhitelist)
		}
	}

	return ""
}

func (i *Instance) getFeed(id string) Feed {
	untypedFeed, ok := i.feeds.Get(id)
	if !ok {
		panic("invalid ingest id: " + id)
	}

	feed, ok := untypedFeed.(Feed)
	if !ok {
		panic("invalid feed type")
	}

	return feed
}

func (i *Instance) getFeedWithWhitelist(id string) FeedWithWhitelist {
	untypedFeedWithWhitelist, ok := i.feeds.Get(id)
	if !ok {
		panic("invalid ingest id: " + id)
	}

	feedWithWhitelist, ok := untypedFeedWithWhitelist.(FeedWithWhitelist)
	if !ok {
		panic("invalid feed with whitelist type")
	}

	return feedWithWhitelist
}

// GetFeedValue returns the most recently published value of a feed along with a string
// representation of the value's type and boolean indicating whether the value is
// okay for consumption.
func (i *Instance) GetFeedValue(id string) (value feed.Value, valueType string, consumable bool) {
	return i.getFeed(id).Value()
}

// GetFeedDefinitions returns a JSON string representing the feed definitions for which the given
// agent address is whitelisted to provide values for ingestion.
func (i *Instance) GetFeedDefinitions(forAddress string) string {
	instanceHasAddressWhitelisted := !i.whitelist.HasDefinition() || i.whitelist.HasAddress(forAddress)

	buf := new(strings.Builder)
	buf.WriteString("[")
	first := true

	i.feeds.Iterate("", "", func(_ string, value interface{}) bool {
		feedWithWhitelist, ok := value.(FeedWithWhitelist)
		if !ok {
			panic("invalid feed type")
		}

		// Don't give agents the ability to try to publish to inactive feeds.
		if !feedWithWhitelist.IsActive() {
			return true
		}

		// Skip feeds the address is not whitelisted for.
		if !addressIsWhitelisted(&i.whitelist, feedWithWhitelist, forAddress, &instanceHasAddressWhitelisted) {
			return true
		}

		taskBytes, err := feedWithWhitelist.Feed.MarshalJSON()
		if err != nil {
			panic(err)
		}

		// Guard against any tasks that shouldn't be returned; maybe they are not active because they have
		// already been completed.
		if len(taskBytes) == 0 {
			return true
		}

		if !first {
			buf.WriteString(",")
		}

		first = false
		buf.Write(taskBytes)
		return true
	})
	buf.WriteString("]")
	return buf.String()
}
