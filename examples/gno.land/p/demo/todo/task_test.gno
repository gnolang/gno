package todo

import (
  "testing"
)

func TestNewTask(t *testing.T) {
  const (
    id     uint   = 1
    task   string = "Debug the code!"
    isDone bool   = false
  )

  j := NewTask(id, task, isDone)

  if (j.GetId() != id) {
    t.Fatalf("Wrong ID!")
  }

  if (j.GetTask() != task) {
    t.Fatalf("Wrong Task!")
  }

  if (j.GetIsDone() != isDone) {
    t.Fatalf("Wrong Status!")
  }
}