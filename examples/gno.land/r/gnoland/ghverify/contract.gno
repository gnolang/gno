package ghverify

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/gnorkle/v1/feed"
	"gno.land/p/demo/gnorkle/v1/feeds/static"
	"gno.land/p/demo/gnorkle/v1/gnorkle"
	"gno.land/p/demo/gnorkle/v1/message"
)

const (
	verifiedResult = "OK"
)

var (
	ownerAddress = string(std.GetOrigCaller())
	oracle       *gnorkle.Instance
	postHandler  postGnorkleMessageHandler

	handleToAddressMap = avl.NewTree()
	addressToHandleMap = avl.NewTree()
)

func init() {
	oracle = gnorkle.NewInstance()
	oracle.AddToWhitelist("", []string{ownerAddress})
}

type postGnorkleMessageHandler struct{}

func (h postGnorkleMessageHandler) Handle(i *gnorkle.Instance, funcType message.FuncType, feed gnorkle.Feed) {
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
			[]feed.Task{NewVerificationTask(string(std.GetOrigCaller()), githubHandle)},
		),
	)
}

func GnorkleEntrypoint(message string) string {
	return oracle.HandleMessage(message, postHandler)
}

func SetOwner(owner string) {
	if ownerAddress != string(std.GetOrigCaller()) {
		panic("only the owner can set a new owner")
	}

	ownerAddress = owner

	// In the context of this contract, the owner is the only one that can
	// add new feeds to the oracle.
	oracle.ClearWhitelist("")
	oracle.AddToWhitelist("", []string{ownerAddress})
}

func Render(_ string) string {
	result := "{"
	var appendComma bool
	handleToAddressMap.Iterate("", "", func(handle string, address interface{}) bool {
		if appendComma {
			result += ","
		}

		result += `"` + handle + `": "` + address.(string) + `"`
		appendComma = true

		return true
	})

	return result + "}"
}
