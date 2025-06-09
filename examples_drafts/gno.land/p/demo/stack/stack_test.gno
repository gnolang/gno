package stack

import "testing"

func TestStack(t *testing.T) {
	s := New() // Empty stack

	if s.Len() != 0 {
		t.Errorf("s.Len(): expected 0; got %d", s.Len())
	}

	s.Push(1)

	if s.Len() != 1 {
		t.Errorf("s.Len(): expected 1; got %d", s.Len())
	}

	if top := s.Top(); top.(int) != 1 {
		t.Errorf("s.Top(): expected 1; got %v", top.(int))
	}

	if elem := s.Pop(); elem.(int) != 1 {
		t.Errorf("s.Pop(): expected 1; got %v", elem.(int))
	}
	if s.Len() != 0 {
		t.Errorf("s.Len(): expected 0; got %d", s.Len())
	}
}
