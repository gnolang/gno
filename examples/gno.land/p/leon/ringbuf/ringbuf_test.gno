package ringbuf

import (
	"testing"
)

func TestRingBufferInt(t *testing.T) {
	buf, err := New(3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	buf.Write(1)
	buf.Write(2)
	buf.Write(3)

	println(buf.PrintBufferData())
	for i := 0; i < buf.size; i++ {
		//val, err := buf.Read()
		//if err != nil || val != i {
		//	t.Errorf("expected 1, got %v (error: %v)", val, err)
		//}
	}

}

func TestRingBufferString(t *testing.T) {
	buf, _ := New(2)
	buf.Write("hello")
	buf.Write("world")

	val, _ := buf.Read()
	if val != "hello" {
		t.Errorf("expected 'hello', got %v", val)
	}
}

func TestRingBufferComplexType(t *testing.T) {
	type ComplexType struct {
		ID   int
		Name string
	}

	buf, _ := New(2)
	buf.Write(ComplexType{ID: 1, Name: "Alice"})
	buf.Write(ComplexType{ID: 2, Name: "Bob"})

	val, _ := buf.Read()
	if v, ok := val.(ComplexType); !ok || v.ID != 1 || v.Name != "Alice" {
		t.Errorf("expected {ID:1, Name: 'Alice'}, got %v", val)
	}
}

func TestRingBufferOverwrite(t *testing.T) {
	buf, _ := New(2)
	buf.Write("A")
	buf.Write("B")
	buf.Write("C") // Overwrites "A"

	val, _ := buf.Read()
	if val != "B" {
		t.Errorf("expected 'B', got %v", val)
	}
}

func TestRingBufferEmptyError(t *testing.T) {
	buf, _ := New(2)

	if _, err := buf.Read(); err != ErrEmpty {
		t.Errorf("expected 'buffer is empty' error, got %v", err)
	}
}
