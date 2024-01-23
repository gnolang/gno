package orkle

import (
	"strings"

	"github.com/gnolang/gno/agent2/p/orkle/agent"
	"github.com/gnolang/gno/agent2/p/orkle/feed"
	"github.com/gnolang/gno/agent2/p/orkle/message"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/std"
)

type Instance struct {
	feeds     *avl.Tree
	whitelist agent.Whitelist
}

func NewInstance() *Instance {
	return &Instance{
		feeds: avl.NewTree(),
	}
}

func assertNonEmptyString(s string) {
	if len(s) == 0 {
		panic("feed ids cannot be empty")
	}
}

func (i *Instance) AddFeeds(feeds ...Feed) {
	for _, feed := range feeds {
		assertNonEmptyString(feed.ID())
		i.feeds.Set(
			feed.ID(),
			FeedWithWhitelist{Feed: feed},
		)
	}
}

func (i *Instance) AddFeedsWithWhitelists(feeds ...FeedWithWhitelist) {
	for _, feed := range feeds {
		assertNonEmptyString(feed.ID())
		i.feeds.Set(
			feed.ID(),
			FeedWithWhitelist{
				Whitelist: feed.Whitelist,
				Feed:      feed,
			},
		)
	}
}

func (i *Instance) RemoveFeed(id string) {
	i.feeds.Remove(id)
}

type PostMessageHandler interface {
	Handle(i *Instance, funcType message.FuncType, feed Feed)
}

func (i *Instance) HandleMessage(msg string, postHandler PostMessageHandler) string {
	caller := string(std.GetOrigCaller())

	funcType, msg := message.ParseFunc(msg)

	switch funcType {
	case message.FuncTypeRequest:
		return i.GetFeedDefinitions(caller)

	default:
		id, msg := message.ParseID(msg)
		feedWithWhitelist := i.getFeedWithWhitelist(id)

		if addressIsWhitelisted(&i.whitelist, feedWithWhitelist, caller, nil) {
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

func (i *Instance) GetFeedValue(id string) (feed.Value, string, bool) {
	return i.getFeed(id).Value()
}

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
