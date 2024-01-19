package orkle

import (
	"strings"

	"github.com/gnolang/gno/agent2/p/orkle/feed"
	"github.com/gnolang/gno/agent2/p/orkle/message"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/std"
)

type Instance struct {
	feeds        *avl.Tree
	whitelist    *avl.Tree
	ownerAddress string
}

func NewInstance(ownerAddress string) *Instance {
	return &Instance{
		ownerAddress: ownerAddress,
	}
}

func (i *Instance) WithWhitelist(addresses ...string) *Instance {
	i.whitelist = avl.NewTree()
	for _, address := range addresses {
		i.whitelist.Set(address, struct{}{})
	}
	return i
}

func (i *Instance) WithFeeds(feeds ...Feed) *Instance {
	i.feeds = avl.NewTree()
	for _, feed := range feeds {
		i.feeds.Set(feed.ID(), feed)
	}
	return i
}

func (i *Instance) AddFeeds(feeds ...Feed) {
	for _, feed := range feeds {
		i.feeds.Set(feed.ID(), feed)
	}
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
		feed := i.getFeed(id)

		if !addressWhitelisted(i.hasAddressWhitelisted(caller), caller, feed) {
			panic("caller not whitelisted")
		}

		feed.Ingest(funcType, msg, caller)

		if postHandler != nil {
			postHandler.Handle(i, funcType, feed)
		}
	}

	return ""
}

func (i *Instance) RemoveFeed(id string) {
	i.feeds.Remove(id)
}

// TODO: test this.

// addressWhiteListed returns true if:
// - the feed has a white list and the address is whitelisted, or
// - the feed has no white list and the instance has a white list and the address is whitelisted, or
// - the feed has no white list and the instance has no white list.
func addressWhitelisted(isInstanceWhitelisted bool, address string, feed Feed) bool {
	// A feed whitelist takes priority, so it will return false if the feed has a whitelist and the caller is
	// not a part of it. An empty whitelist defers to the instance whitelist.
	var isWhitelisted, hasWhitelist bool
	if isWhitelisted, hasWhitelist = feed.HasAddressWhitelisted(address); !isWhitelisted && hasWhitelist {
		return false
	}

	return (isWhitelisted && hasWhitelist) || isInstanceWhitelisted
}

// hasAddressWhitelisted returns true if the address is whitelisted for the instance or if the instance has
// no whitelist.
func (i *Instance) hasAddressWhitelisted(address string) bool {
	return i.whitelist == nil || i.whitelist.Has(address)
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

func (i *Instance) GetFeedValue(id string) (feed.Value, string, bool) {
	return i.getFeed(id).Value()
}

func (i *Instance) GetFeedDefinitions(forAddress string) string {
	instanceHasAddressWhitelisted := i.hasAddressWhitelisted(forAddress)

	buf := new(strings.Builder)
	buf.WriteString("[")
	first := true

	i.feeds.Iterate("", "", func(_ string, value interface{}) bool {
		feed, ok := value.(Feed)
		if !ok {
			panic("invalid feed type")
		}

		// Don't give agents the ability to try to publish to inactive feeds.
		if !feed.IsActive() {
			return true
		}

		// Skip feeds the address is not whitelisted for.
		if !addressWhitelisted(instanceHasAddressWhitelisted, forAddress, feed) {
			return true
		}

		taskBytes, err := feed.MarshalJSON()
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
