package helloworld

var message string = "Hello, World!"

func SetMessage(cur realm, newMessage string) {
	message = newMessage
}

func RenderDemo() string {
	content := "# DEMO\n"
	content += "Message Stored : **" + message + "**\n\n"
	content += "[SetMessage](helloworld$help&func=SetMessage)\n\n"
	return content
}
