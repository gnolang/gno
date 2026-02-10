package ghverify

import (
	"chain"
	"chain/runtime"
	"errors"

	"gno.land/p/demo/gnorkle/feeds/static"
	"gno.land/p/demo/gnorkle/gnorkle"
	"gno.land/p/demo/gnorkle/message"
	"gno.land/p/nt/avl"
)

const (
	// The agent should send this value if it has verified the github handle.
	verifiedResult = "OK"
)

var (
	ownerAddress = runtime.OriginCaller()
	oracle       *gnorkle.Instance
	postHandler  postGnorkleMessageHandler

	handleToAddressMap = avl.NewTree()
	addressToHandleMap = avl.NewTree()
)

func init() {
	oracle = gnorkle.NewInstance()
	oracle.AddToWhitelist("", []string{string(ownerAddress)})
}

type postGnorkleMessageHandler struct{}

// Handle does post processing after a message is ingested by the oracle feed. It extracts the value to realm
// storage and removes the feed from the oracle.
func (h postGnorkleMessageHandler) Handle(i *gnorkle.Instance, funcType message.FuncType, feed gnorkle.Feed) error {
	if funcType != message.FuncTypeIngest {
		return nil
	}

	result, _, consumable := feed.Value()
	if !consumable {
		return nil
	}

	// The value is consumable, meaning the ingestion occurred, so we can remove the feed from the oracle
	// after saving it to realm storage.
	defer oracle.RemoveFeed(feed.ID())

	// Couldn't verify; nothing to do.
	if result.String != verifiedResult {
		return nil
	}

	feedTasks := feed.Tasks()
	if len(feedTasks) != 1 {
		return errors.New("expected feed to have exactly one task")
	}

	task, ok := feedTasks[0].(*verificationTask)
	if !ok {
		return errors.New("expected ghverify task")
	}

	handleToAddressMap.Set(task.githubHandle, task.gnoAddress)
	addressToHandleMap.Set(task.gnoAddress, task.githubHandle)
	return nil
}

// RequestVerification creates a new static feed with a single task that will
// instruct an agent to verify the github handle / gno address pair.
func RequestVerification(cur realm, githubHandle string) {
	gnoAddress := string(runtime.OriginCaller())
	if err := oracle.AddFeeds(
		static.NewSingleValueFeed(
			gnoAddress,
			"string",
			&verificationTask{
				gnoAddress:   gnoAddress,
				githubHandle: githubHandle,
			},
		),
	); err != nil {
		panic(err)
	}
	chain.Emit(
		"verification_requested",
		"from", gnoAddress,
		"handle", githubHandle,
	)
}

// GnorkleEntrypoint is the entrypoint to the gnorkle oracle handler.
func GnorkleEntrypoint(cur realm, message string) string {
	result, err := oracle.HandleMessage(message, postHandler)
	if err != nil {
		panic(err)
	}

	return result
}

// SetOwner transfers ownership of the contract to the given address.
func SetOwner(_ realm, owner address) {
	if ownerAddress != runtime.OriginCaller() {
		panic("only the owner can set a new owner")
	}

	ownerAddress = owner

	// In the context of this contract, the owner is the only one that can
	// add new feeds to the oracle.
	oracle.ClearWhitelist("")
	oracle.AddToWhitelist("", []string{string(ownerAddress)})
}

// GetHandleByAddress returns the github handle associated with the given gno address.
func GetHandleByAddress(cur realm, address_XXX string) string {
	if value, ok := addressToHandleMap.Get(address_XXX); ok {
		return value.(string)
	}

	return ""
}

// GetAddressByHandle returns the gno address associated with the given github handle.
func GetAddressByHandle(cur realm, handle string) string {
	if value, ok := handleToAddressMap.Get(handle); ok {
		return value.(string)
	}

	return ""
}

// Render returns a json object string will all verified handle -> address mappings.
func Render(_ string) string {
	result := "{"
	var appendComma bool
	handleToAddressMap.Iterate("", "", func(handle string, address_XXX any) bool {
		if appendComma {
			result += ","
		}

		result += `"` + handle + `": "` + address_XXX.(string) + `"`
		appendComma = true

		return false
	})

	return result + "}"
}
