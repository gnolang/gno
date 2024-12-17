package dao_interfaces

import (
	"testing"

	"gno.land/p/demo/json"
)

func TestRegistry(t *testing.T) {
	registry := NewMessagesRegistry()

	var value string
	msgHandler := NewCopyMessageHandler(&value)

	// Test register handler via message
	registerMsg := &RegisterHandlerExecutableMessage{Handler: msgHandler}
	registry.Execute(registerMsg)

	// Test messages execution
	msgs := registry.MessagesFromJSON(json.Must(json.Unmarshal([]byte(`[{"type":"CopyMessage","payload":"Hello"}]`))).MustArray())
	if len(msgs) != 1 {
		t.Errorf("Expected 1 message, got %d", len(msgs))
	}

	registry.Execute(msgs[0])
	if value != "Hello" {
		t.Errorf("Expected value to be 'Hello', got '%s'", value)
	}

	msg2 := &CopyMessage{Value: "World"}
	registry.Execute(msg2)
	if value != "World" {
		t.Errorf("Expected value to be 'World', got '%s'", value)
	}

	// Test handler removal
	removeMsg := &RemoveHandlerExecutableMessage{HandlerType: msgHandler.Type()}
	registry.Execute(removeMsg)
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic, got none")
			}
		}()
		registry.Execute(msg2)
	}()

	// Test direct register
	registry.Register(msgHandler)
	msg3 := &CopyMessage{Value: "!"}
	registry.Execute(msg3)
	if value != "!" {
		t.Errorf("Expected value to be '!', got '%s'", value)
	}
}
