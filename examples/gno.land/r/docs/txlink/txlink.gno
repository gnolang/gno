package txlink

import (
	"chain/runtime"
	"time"

	"gno.land/p/moul/txlink"
)

type Message struct {
	Author string
	Text   string
	Time   time.Time
}

var messages []Message

// Post saves a message with the caller as the author
func Post(_ realm, text string) {
	msg := Message{
		Author: runtime.PreviousRealm().Address().String(),
		Text:   text,
		Time:   time.Now(),
	}
	messages = append(messages, msg)
}

func Render(_ string) string {
	out := "# Transaction Links\n\n"
	out += `Transaction links ("txlinks" for short) are a standardized way of creating
URLs that Gno-enabled wallets can interpret and generate transactions from.

They allow developers to turn basic hyperlinks into realm calls
with optional arguments and funds attached.

Arguments with empty values are treated as _wallet-settable_, allowing the
user to customize them before submitting the transaction.

You can also link to functions in _other realms_, using the pkgpath of the realm in question.

Check out the links below.

`

	out += "## Post a Message\n"

	// Basic call
	out += "- [Say Hello](" + txlink.Call("Post", "text", "Hello from txlink!") + ") - A simple, static argument call\n"

	// A link builder including some funds specified in the string format (std.Coin.String())
	custom := txlink.NewLink("Post").
		AddArgs("text", "Support this board with 1ugnot").
		SetSend("1ugnot").
		URL()
	out += "- [Send Supportive Message](" + custom + ") - A more complex call with coins included\n"

	// Up to the user to set the argument
	out += "- [Write Your Own Message](" + txlink.NewLink("Post").
		AddArgs("text", ""). // Should be editable in the user's wallet
		URL() + ") - A user-defined argument call\n"

	// Cross-realm example
	crossRealm := txlink.Realm("gno.land/r/docs/minisocial/v2").
		NewLink("CreatePost").
		AddArgs("text", "Hello from `/r/docs/txlink`!").
		URL()
	out += "- [Say Hello on MiniSocial V2!](" + crossRealm + ") - A call to another realm\n"

	out += "\n## Recent Messages\n"
	if len(messages) == 0 {
		out += "_No messages yet._\n\n"
		return out
	}

	for i := len(messages) - 1; i >= 0 && i >= len(messages)-5; i-- {
		msg := messages[i]
		out += "- **" + msg.Author + "**: " + msg.Text + " (" + msg.Time.Format("15:04:05") + ")\n"
	}

	return out
}
