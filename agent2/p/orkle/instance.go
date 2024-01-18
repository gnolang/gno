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

type PostMessageHandler interface {
	Handle(i *Instance, funcType message.FuncType, feed Feed)
}

func (i *Instance) HandleMessage(msg string, postHandler PostMessageHandler) string {
	caller := string(std.GetOrigCaller())
	if i.whitelist != nil {
		// Check that the caller is whitelisted.
		if _, ok := i.whitelist.Get(caller); !ok {
			panic("caller not whitelisted")
		}
	}

	funcType, msg := message.ParseFunc(msg)

	switch funcType {
	case message.FuncTypeRequest:
		return i.RequestTasks()

	default:
		id, msg := message.ParseID(msg)
		feed := i.getFeed(id)

		feed.Ingest(funcType, msg, caller)

		if postHandler != nil {
			postHandler.Handle(i, funcType, feed)
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

func (i *Instance) GetFeedValue(id string) (feed.Value, string, bool) {
	return i.getFeed(id).Value()
}

func (i *Instance) RequestTasks() string {
	buf := new(strings.Builder)
	buf.WriteString("[")
	first := true

	i.feeds.Iterate("", "", func(_ string, value interface{}) bool {
		if !first {
			buf.WriteString(",")
		}

		task, ok := value.(Feed)
		if !ok {
			panic("invalid task type")
		}

		taskBytes, err := task.MarshalJSON()
		if err != nil {
			panic(err)
		}

		// Guard against any tasks that shouldn't be returned; maybe they are not active because they have
		// already been completed.
		if len(taskBytes) == 0 {
			return true
		}

		first = false
		buf.Write(taskBytes)
		return true
	})
	buf.WriteString("]")
	return buf.String()
}
