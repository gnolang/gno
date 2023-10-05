package helloworld

import "testing"

func TestSayHelloWorld(t *testing.T) {
	expected := "Hello World!"
	t.Run("Correct", func(t *testing.T) {
		out := SayHelloWorld()
		if out != expected {
			t.Errorf("expected %s got %s", expected, out)
		}
	})
}
