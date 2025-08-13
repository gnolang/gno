package statelock

type State string

const (
	StateActive   State = "Active"
	StateLocked   State = "Locked"
	StateDisabled State = "Disabled"
)

var state State = StateActive

func assertInState(expected State) {
	if state != expected {
		panic("invalid state: expected " + expected + ", got " + state)
	}
}

func Lock(cur realm) {
	assertInState(StateActive)
	state = StateLocked
}

func Unlock(cur realm) {
	assertInState(StateLocked)
	state = StateActive
}

func RenderDemo() string {
	content := "## Live Demo\n"
	content += "Contract State : **" + string(state) + "**\n\n"
	content += "[Lock](statelock$help&func=Lock)\n\n"
	content += "[Unlock](statelock$help&func=Unlock)\n\n"
	return content
}
