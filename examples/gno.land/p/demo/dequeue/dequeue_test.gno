package dequeue

import (
	"testing"
)

func TestPushBack(t *testing.T) {
	// Initialize a new queue
	queue := New()

	// Test pushing an element to the back of the queue
	err := queue.PushBack("first")
	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if queue.Length() != 1 {
		t.Errorf("Expected length 1, got: %d", queue.Length())
	}

	// Test pushing another element to the back of the queue
	err = queue.PushBack("second")
	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if queue.Length() != 2 {
		t.Errorf("Expected length 2, got: %d", queue.Length())
	}
}

func TestPopBack(t *testing.T) {
	// Initialize a new queue
	queue := New()

	// Add two elements to the queue
	queue.PushBack("first")
	queue.PushBack("second")

	// Test popping the last element from the queue
	value, err := queue.PopBack()
	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if value != "second" {
		t.Errorf("Expected 'second', got: %s", value)
	}

	if queue.Length() != 1 {
		t.Errorf("Expected length 1, got: %d", queue.Length())
	}

	// Test popping the last remaining element from the queue
	value, err = queue.PopBack()
	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if value != "first" {
		t.Errorf("Expected 'first', got: %s", value)
	}

	if queue.Length() != 0 {
		t.Errorf("Expected length 0, got: %d", queue.Length())
	}

	// Test popping from an empty queue, should return an error
	_, err = queue.PopBack()
	if err != ErrEmptyArrayPop {
		t.Errorf("Expected error %s, got: %s", ErrEmptyArrayPop, err.Error())
	}
}

func TestPushFront(t *testing.T) {
	// Initialize a new queue
	queue := New()

	// Test pushing an element to the front of the queue
	err := queue.PushFront("first")
	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if queue.Length() != 1 {
		t.Errorf("Expected length 1, got: %d", queue.Length())
	}

	// Test pushing another element to the front of the queue
	err = queue.PushFront("second")
	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if queue.Length() != 2 {
		t.Errorf("Expected length 2, got: %d", queue.Length())
	}
}

func TestPopFront(t *testing.T) {
	// Initialize a new queue
	queue := New()

	// Add two elements to the queue
	queue.PushBack("first")
	queue.PushBack("second")

	// Test popping the front element from the queue
	value, err := queue.PopFront()
	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if value != "first" {
		t.Errorf("Expected 'first', got: %s", value)
	}

	if queue.Length() != 1 {
		t.Errorf("Expected length 1, got: %d", queue.Length())
	}

	// Test popping the next front element from the queue
	value, err = queue.PopFront()
	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if value != "second" {
		t.Errorf("Expected 'second', got: %s", value)
	}

	if queue.Length() != 0 {
		t.Errorf("Expected length 0, got: %d", queue.Length())
	}

	// Test popping from an empty queue, should return an error
	_, err = queue.PopFront()
	if err != ErrEmptyArrayPop {
		t.Errorf("Expected error %s, got: %s", ErrEmptyArrayPop, err.Error())
	}
}

func TestFront(t *testing.T) {
	// Initialize a new queue
	queue := New()

	// Add two elements to the queue
	queue.PushBack("first")
	queue.PushBack("second")

	// Test retrieving the front element without removing it
	value, err := queue.Front()
	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if value != "first" {
		t.Errorf("Expected 'first', got: %s", value)
	}

	// Remove the front element and test retrieving the new front element
	queue.PopFront()
	value, err = queue.Front()
	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if value != "second" {
		t.Errorf("Expected 'second', got: %s", value)
	}

	// Remove the remaining element and test retrieving the front from an empty queue, should return an error
	queue.PopFront()
	_, err = queue.Front()
	if err != ErrArrayOutOfBounds {
		t.Errorf("Expected error %s, got: %s", ErrArrayOutOfBounds, err.Error())
	}
}

func TestBack(t *testing.T) {
	// Initialize a new queue
	queue := New()

	// Add two elements to the queue
	queue.PushBack("first")
	queue.PushBack("second")

	// Test retrieving the last element without removing it
	value, err := queue.Back()
	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if value != "second" {
		t.Errorf("Expected 'second', got: %s", value)
	}

	// Remove the last element and test retrieving the new last element
	queue.PopBack()
	value, err = queue.Back()
	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if value != "first" {
		t.Errorf("Expected 'first', got: %s", value)
	}

	// Remove the remaining element and test retrieving the last from an empty queue, should return an error
	queue.PopBack()
	_, err = queue.Back()
	if err != ErrArrayOutOfBounds {
		t.Errorf("Expected error %s, got: %s", ErrArrayOutOfBounds, err.Error())
	}
}

func TestAt(t *testing.T) {
	// Initialize a new queue
	queue := New()

	// Add three elements to the queue
	queue.PushBack("first")
	queue.PushBack("second")
	queue.PushBack("third")

	// Test accessing elements at specific indices
	value, err := queue.At(0)
	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if value != "first" {
		t.Errorf("Expected 'first', got: %s", value)
	}

	value, err = queue.At(1)
	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if value != "second" {
		t.Errorf("Expected 'second', got: %s", value)
	}

	value, err = queue.At(2)
	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if value != "third" {
		t.Errorf("Expected 'third', got: %s", value)
	}

	// Test accessing an out-of-bounds index, should return an error
	_, err = queue.At(3)
	if err != ErrArrayOutOfBounds {
		t.Errorf("Expected error %s, got: %s", ErrArrayOutOfBounds, err.Error())
	}
}

func TestClear(t *testing.T) {
	// Initialize a new queue
	queue := New()

	// Add two elements to the queue
	queue.PushBack("first")
	queue.PushBack("second")

	// Clear the queue
	queue.Clear()

	// Test if the queue is empty after clearing
	if queue.Length() != 0 {
		t.Errorf("Expected length 0, got: %d", queue.Length())
	}

	if !queue.Empty() {
		t.Errorf("Expected queue to be empty")
	}

	// Test retrieving the front from an empty queue, should return an error
	_, err := queue.Front()
	if err != ErrArrayOutOfBounds {
		t.Errorf("Expected error %s, got: %s", ErrArrayOutOfBounds, err.Error())
	}

	// Test retrieving the back from an empty queue, should return an error
	_, err = queue.Back()
	if err != ErrArrayOutOfBounds {
		t.Errorf("Expected error %s, got: %s", ErrArrayOutOfBounds, err.Error())
	}
}

func TestLengthAndEmpty(t *testing.T) {
	// Initialize a new queue
	queue := New()

	// Test initial length and empty status
	if queue.Length() != 0 {
		t.Errorf("Expected length 0, got: %d", queue.Length())
	}

	if !queue.Empty() {
		t.Errorf("Expected queue to be empty")
	}

	// Add an element and test length and empty status
	queue.PushBack("first")
	if queue.Length() != 1 {
		t.Errorf("Expected length 1, got: %d", queue.Length())
	}

	if queue.Empty() {
		t.Errorf("Expected queue not to be empty")
	}

	// Add another element and test length and empty status
	queue.PushBack("second")
	if queue.Length() != 2 {
		t.Errorf("Expected length 2, got: %d", queue.Length())
	}

	if queue.Empty() {
		t.Errorf("Expected queue not to be empty")
	}

	// Remove an element and test length and empty status
	queue.PopFront()
	if queue.Length() != 1 {
		t.Errorf("Expected length 1, got: %d", queue.Length())
	}

	if queue.Empty() {
		t.Errorf("Expected queue not to be empty")
	}

	// Remove the last element and test length and empty status
	queue.PopBack()
	if queue.Length() != 0 {
		t.Errorf("Expected length 0, got: %d", queue.Length())
	}

	if !queue.Empty() {
		t.Errorf("Expected queue to be empty")
	}
}
