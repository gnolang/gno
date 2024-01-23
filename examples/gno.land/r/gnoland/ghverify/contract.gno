package ghverify

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/orkle/v1/feeds/static"
	"gno.land/p/demo/orkle/v1/message"
	"gno.land/p/demo/orkle/v1/orkle"
)

const (
	verifiedResult          = "OK"
	whitelistedAgentAddress = "..."
)

var (
	oracle      *orkle.Instance
	postHandler postOrkleMessageHandler

	handleToAddressMap = avl.NewTree()
	addressToHandleMap = avl.NewTree()
)

func init() {
	oracle = orkle.NewInstance()
	oracle.AddToWhitelist("", []string{whitelistedAgentAddress})
}

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

	task, ok := feedTasks[0].(*verificationTask)
	if !ok {
		panic("expected ghverify task")
	}

	handleToAddressMap.Set(task.GithubHandle(), task.GnoAddress())
	addressToHandleMap.Set(task.GnoAddress(), task.GithubHandle())
}

func RequestVerification(githubHandle string) {
	oracle.AddFeeds(
		static.NewSingleValueFeed(
			githubHandle,
			"string",
			nil,
			NewVerificationTask(string(std.GetOrigCaller()), githubHandle),
		),
	)
}

func OrkleEntrypoint(message string) string {
	return oracle.HandleMessage(message, postHandler)
}
