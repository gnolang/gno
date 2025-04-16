package organizer

import "testing"

func Test_getGroupFromId_invalidID(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Invalid ID" {
			t.Errorf("Expected Invalid ID panic")
		}
	}()

	getGroupFromId("abc")
}
