package statelock

type State string

const (
	StateActive State = "Active"
	StateLocked State = "Locked"
	// StateDisabled mirrors Solidity's three-value enum; no transition uses it in this demo.
	StateDisabled State = "Disabled"
)

var state State = StateActive

func assertInState(expected State) {
	if state != expected {
		panic("invalid state: expected " + string(expected) + ", got " + string(state))
	}
}

// Lock transitions the contract from Active to Locked.
// Any authenticated caller may lock.
func Lock(cur realm) {
	assertInState(StateActive)
	state = StateLocked
}

// Unlock transitions the contract from Locked to Active.
// Same caller model as Lock: any authenticated caller may unlock.
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
