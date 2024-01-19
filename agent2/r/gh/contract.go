package gh

import (
	"github.com/gnolang/gno/agent2/p/orkle"
	"github.com/gnolang/gno/agent2/p/orkle/feed/static"
	"github.com/gnolang/gno/agent2/p/orkle/feed/tasks/ghverify"
	"github.com/gnolang/gno/agent2/p/orkle/message"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/std"
)

const (
	verifiedResult          = "OK"
	whitelistedAgentAddress = "..."
)

var (
	oracle      orkle.Instance
	postHandler postOrkleMessageHandler

	handleToAddressMap = avl.NewTree()
	addressToHandleMap = avl.NewTree()
)

type postOrkleMessageHandler struct{}

func (h postOrkleMessageHandler) Handle(i *orkle.Instance, funcType message.FuncType, feed orkle.Feed) {
	if funcType != message.FuncTypeIngest {
		return
	}

	result, _, consumable := feed.Value()
	if !consumable {
		return
	}

	defer oracle.RemoveFeed(feed.ID())

	if result.String != verifiedResult {
		return
	}

	feedTasks := feed.Tasks()
	if len(feedTasks) != 1 {
		panic("expected feed to have exactly one task")
	}

	task, ok := feedTasks[0].(*ghverify.Task)
	if !ok {
		panic("expected ghverify task")
	}

	handleToAddressMap.Set(task.GithubHandle(), task.GnoAddress())
	addressToHandleMap.Set(task.GnoAddress(), task.GithubHandle())
}

func init() {
	oracle = *orkle.NewInstance(string(std.GetOrigCaller())).
		WithWhitelist(whitelistedAgentAddress)
}

func RequestVerification(githubHandle string) {
	oracle.AddFeeds(
		static.NewSingleValueFeed(
			githubHandle,
			"string",
			nil,
			ghverify.NewTask(string(std.GetOrigCaller()), githubHandle),
		),
	)
}

func OrkleEntrypoint(message string) string {
	return oracle.HandleMessage(message, postHandler)
}
